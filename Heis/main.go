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
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

// Test the prosess pairs by using the commands:
// ps aux | grep main
// kill -SIGINT <PID>
const (
	localPort         = "30210"
	backupPort        = "30211"         // Separate port for backup
	readTimeout       = 2 * time.Second // Increased timeout for robustness
	heartbeatInterval = 500 * time.Millisecond
)

func main() {
	// Command-line flags for role and configuration
	var id, port, role string
	flag.StringVar(&id, "id", "", "ID of this peer")
	flag.StringVar(&port, "port", "15657", "Elevator IO port")
	flag.StringVar(&role, "role", "primary", "Role: primary or backup")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// Decide behavior based on role
	if role == "primary" {
		runPrimary(id, port)
	} else {
		runBackup(id, port)
	}
}

func runPrimary(id, port string) {
	fmt.Println("Starting as primary...")

	// Dial backup for heartbeats
	backupAddr, err := net.ResolveUDPAddr("udp", "localhost:"+backupPort)
	if err != nil {
		log.Fatalf("Failed to resolve backup address: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, backupAddr)
	if err != nil {
		log.Fatalf("Failed to dial backup: %v", err)
	}
	defer conn.Close()

	// Spawn backup process
	spawnNewInstance("main.go", port, "backup")

	// Start elevator system
	startElevatorSystem(id, port)

	// Send heartbeats
	go sendHeartbeat(conn)

	// Handle shutdown
	waitForShutdown()
}

func runBackup(id, port string) {
	fmt.Println("Starting as backup...")

	// Listen for heartbeats
	udpAddr, err := net.ResolveUDPAddr("udp", ":"+backupPort)
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to start UDP listener: %v", err)
	}
	defer conn.Close()

	// Detect primary failure
	detectFailure(conn)

	// Take over as primary
	fmt.Println("Primary failed, taking over...")
	runPrimary(id, port) // Restart as primary
}

func startElevatorSystem(id, port string) {
	// Load configuration
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize elevator IO
	elevio.Init("localhost:"+port, cfg.NumFloors)

	// Channels
	peerUpdateCh := make(chan peers.PeerUpdate)
	remoteElevatorCh := make(chan msgTypes.ElevatorStateMsg)
	peerTxEnable := make(chan bool)
	localAssignedOrderCH := make(chan elevio.ButtonEvent)
	buttonPressCH := make(chan elevio.ButtonEvent)
	completedOrderCh := make(chan elevio.ButtonEvent, 3)
	fsmToOrdersCH := make(chan elevator.Elevator)
	ordersToPeersCH := make(chan elevator.NetworkElevator)
	networkDisconnectCh := make(chan bool)

	// Start goroutines
	go peers.Transmitter(17135, id, peerTxEnable, networkDisconnectCh, ordersToPeersCH)
	go peers.Receiver(17135, id, peerUpdateCh, remoteElevatorCh)
	go fsm.Fsm(id, localAssignedOrderCH, buttonPressCH, completedOrderCh, fsmToOrdersCH)
	go orders.OrderHandler(id, localAssignedOrderCH, buttonPressCH, completedOrderCh, remoteElevatorCh, peerUpdateCh, networkDisconnectCh, fsmToOrdersCH, ordersToPeersCH)
}

func detectFailure(conn *net.UDPConn) {
	buffer := make([]byte, 1024)
	missedHeartbeats := 0
	const maxMissed = 1

	for {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		n, addr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				missedHeartbeats++
				log.Printf("Missed heartbeat %d/%d", missedHeartbeats, maxMissed)
				if missedHeartbeats >= maxMissed {
					log.Println("Primary process not responding. Taking over...")
					return
				}
				continue
			}
			log.Printf("Error reading from UDP: %v", err)
			continue
		}

		if string(buffer[:n]) == "heartbeat" {
			log.Printf("Received heartbeat from %v", addr)
			missedHeartbeats = 0 // Reset on successful heartbeat
		}
	}
}

func spawnNewInstance(fileName, port, role string) {
	cmd := exec.Command("go", "run", fileName, "-port="+port, "-role="+role)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start new instance: %v", err)
	}
	log.Println("New process started successfully.")

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Printf("Process exited with error: %v", err)
		}
	}()
}

func sendHeartbeat(conn *net.UDPConn) {
	for {
		_, err := conn.Write([]byte("heartbeat"))
		if err != nil {
			log.Printf("Failed to send heartbeat: %v", err)
		}
		time.Sleep(heartbeatInterval)
	}
}

func waitForShutdown() {
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Shutting down...")
}
