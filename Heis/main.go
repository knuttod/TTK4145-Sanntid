package main

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/localip"
	"Heis/pkg/network/peers"
	"Heis/pkg/orders"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

func main() {
	var id string
	var port string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.StringVar(&port, "port", "", "port of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	backupPort := "30210"
	addr := ":" + backupPort

	s, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("Error resolving UDP address: %v", err)
		return
	}

	connection, err := net.ListenUDP("udp", s)
	if err != nil {
		fmt.Println("Starting as primary...")
		go runAsPrimary(id, port)

		go func() {
			time.Sleep(1 * time.Second)
			err := exec.Command("gnome-terminal", "--", "go", "run", "main.go", "-id", id+"-backup", "-port", port).Run()
			if err != nil {
				log.Printf("Failed to spawn backup: %v", err)
			}
		}()
	} else {
		fmt.Println("Starting as backup...")
		runAsBackup(id, port, connection)
	}

	select {}
}

func runAsPrimary(id string, port string) {
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	elevio.Init("localhost:"+port, cfg.NumFloors)

	peerUpdateCh := make(chan peers.PeerUpdate)
	remoteElevatorCh := make(chan msgTypes.ElevatorStateMsg)
	peerTxEnable := make(chan bool)
	localAssignedOrderCH := make(chan elevio.ButtonEvent)
	buttonPressCH := make(chan elevio.ButtonEvent)
	completedOrderCh := make(chan elevio.ButtonEvent, 3)
	fsmToOrdersCH := make(chan elevator.Elevator)
	ordersToPeersCH := make(chan elevator.NetworkElevator)
	networkDisconnectCh := make(chan bool) // Fixed typo
	transmitterToRecivierSkipCh := make(chan bool)

	// UDP connection to backup
	backupAddr, err := net.ResolveUDPAddr("udp", ":30210")
	if err != nil {
		log.Printf("Failed to resolve backup address: %v", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, backupAddr)
	if err != nil {
		log.Printf("Failed to dial backup: %v", err)
		return
	}

	// Assuming peers.Receiver expects channels as input-only (chan<-)
	go peers.Transmitter(17135, id, peerTxEnable, transmitterToRecivierSkipCh, ordersToPeersCH)
	go peers.Receiver(17135, id, peerUpdateCh, remoteElevatorCh) // Adjusted arguments
	go fsm.Fsm(id, localAssignedOrderCH, buttonPressCH, completedOrderCh, fsmToOrdersCH)
	go orders.OrderHandler(id, localAssignedOrderCH, buttonPressCH, completedOrderCh,
		remoteElevatorCh, peerUpdateCh, networkDisconnectCh, fsmToOrdersCH, ordersToPeersCH) // Fixed variable name

	// Periodic state sync to backup
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			// Using conn to avoid "declared and not used" error
			_, err := conn.Write([]byte("ping")) // Simple heartbeat
			if err != nil {
				log.Printf("Failed to send to backup: %v", err)
			}
		}
	}()
}

func runAsBackup(id string, port string, connection *net.UDPConn) {
	defer connection.Close()

	buffer := make([]byte, 1024)

	for {
		connection.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := connection.ReadFromUDP(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("Primary failed, taking over as primary...")
				go func() {
					err := exec.Command("gnome-terminal", "--", "go", "run", "main.go",
						"-id", id+"-backup", "-port", port).Run()
					if err != nil {
						log.Printf("Failed to spawn new backup: %v", err)
					}
				}()
				go runAsPrimary(id, port)
				return
			}
		}
	}
}
