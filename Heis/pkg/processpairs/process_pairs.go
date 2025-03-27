package processpairs

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/network/network"
	"Heis/pkg/orders"
	"log"
	"net"
	"os/exec"
	"time"
)

// PrimarySetup configures and runs the elevator system as the primary process
func PrimarySetup(id string, port string, backupPort string, conn *net.UDPConn) {
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	elevio.Init("localhost:"+port, cfg.NumFloors)

	// between fsm and orders
	buttonPressCH := make(chan elevio.ButtonEvent)
	completedOrderCh := make(chan elevio.ButtonEvent)
	fsmToOrdersCH := make(chan elevator.Elevator)
	localAssignedOrderCh := make(chan elevio.ButtonEvent)

	// between orders and network
	ordersToPeersCH := make(chan elevator.NetworkElevator)
	peerUpdateCh := make(chan network.PeerUpdate)
	remoteElevatorUpdateCh := make(chan network.ElevatorStateMsg)

	// enable sending on network
	peerTxEnable := make(chan bool)
	// between transmitter and reciever
	transmitterToRecivierSkipCh := make(chan bool)

	// Launch main elevator system components as goroutines
	go fsm.Fsm(id, localAssignedOrderCh, buttonPressCH, completedOrderCh, fsmToOrdersCH)
	go orders.OrderHandler(id, localAssignedOrderCh, buttonPressCH, completedOrderCh, remoteElevatorUpdateCh, peerUpdateCh, fsmToOrdersCH, ordersToPeersCH)
	go network.Transmitter(17135, id, peerTxEnable, transmitterToRecivierSkipCh, ordersToPeersCH)
	go network.Receiver(17135, id, transmitterToRecivierSkipCh, peerUpdateCh, remoteElevatorUpdateCh)

	// Periodic heartbeat to backup
	go heartBeat(conn)
}


func heartBeat(conn *net.UDPConn) {
	for {
		time.Sleep(500 * time.Millisecond)
		_, err := conn.Write([]byte("ping")) 
		if err != nil {
			log.Printf("Failed to send to backup: %v", err)
		}
	}
}

// BackupSetup configures and runs the elevator system as the backup process
func BackupSetup(id string, port string, connection *net.UDPConn, backupPort string) {
	defer connection.Close()

	buffer := make([]byte, 1024)

	for {
		connection.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, _, err := connection.ReadFromUDP(buffer)

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Println("Primary failed, taking over as primary...")
				return
			}
		}
	}
}

// SetupUDPListener attempts to create a UDP listener to determine role
func SetupUDPListener(backupPort string) (*net.UDPConn, error) {
	addr := ":" + backupPort
	s, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		log.Printf("Error resolving UDP address: %v", err)
		return nil, err
	}
	return net.ListenUDP("udp", s)
}

// StartPrimaryProcess initializes the primary process with UDP connection
func StartPrimaryProcess(id, port, backupPort string) {
	backupAddr, err := net.ResolveUDPAddr("udp", ":"+backupPort)
	if err != nil {
		log.Printf("Failed to resolve backup address: %v", err)
		return
	}
	conn, err := net.DialUDP("udp", nil, backupAddr)
	if err != nil {
		log.Printf("Failed to dial backup: %v", err)
		return
	}
	PrimarySetup(id, port, backupPort, conn)
}

// SpawnBackupProcess launches the backup process in a new terminal
func SpawnBackupProcess(id, port string) {
	time.Sleep(1 * time.Second)
	err := exec.Command("gnome-terminal", "--", "go", "run", "main.go",
		"-id", id+"-backup", "-port", port, "-processPairsFlag=true").Run()
	if err != nil {
		log.Printf("Failed to spawn backup: %v", err)
	}
}

// MonitorAndTakeOver handles backup taking over as primary when needed
func MonitorAndTakeOver(id, port string, connection *net.UDPConn, backupPort string) {
	BackupSetup(id, port, connection, backupPort)
	go func() {
		err := exec.Command("gnome-terminal", "--", "go", "run", "main.go",
			"-id", id, "-port", port, "-processPairsFlag=true").Run()
		if err != nil {
			log.Printf("Failed to spawn new backup: %v", err)
		}
	}()
	go StartPrimaryProcess(id, port, backupPort)
}
