package processPairs

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/elevio"
	"Heis/pkg/fsm"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/peers"
	"Heis/pkg/orders"
	"fmt"
	"log"
	"net"
	"time"
)

// RunPrimary configures and runs the elevator system as the primary process
func RunPrimary(cfg *Config) error {
	// Load elevator configuration
	elevCfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Initialize hardware
	elevio.Init("localhost:"+cfg.Port, elevCfg.NumFloors)

	// Channels for inter-module communication
	channels := NewChannels()

	// UDP connection to backup
	conn, err := setupBackupConnection(cfg.BackupPort)
	if err != nil {
		log.Printf("Failed to setup backup connection: %v", err)
	}

	// Launch system components
	go peers.Transmitter(17135, cfg.ID, channels.PeerTxEnable, channels.SkipCh, channels.OrdersToPeers)
	go peers.Receiver(17135, cfg.ID, channels.SkipCh, channels.PeerUpdate, channels.RemoteElevator)
	go fsm.Fsm(cfg.ID, channels.LocalAssignedOrder, channels.ButtonPress, channels.CompletedOrder, channels.FsmToOrders)
	go orders.OrderHandler(cfg.ID, channels.LocalAssignedOrder, channels.ButtonPress, channels.CompletedOrder,
		channels.RemoteElevator, channels.PeerUpdate, channels.NetworkDisconnect, channels.FsmToOrders, channels.OrdersToPeers)

	// Periodic state sync to backup
	go syncWithBackup(conn)

	return nil
}

// Channels encapsulates all communication channels
type Channels struct {
	PeerUpdate         chan peers.PeerUpdate
	RemoteElevator     chan msgTypes.ElevatorStateMsg
	PeerTxEnable       chan bool
	LocalAssignedOrder chan elevio.ButtonEvent
	ButtonPress        chan elevio.ButtonEvent
	CompletedOrder     chan elevio.ButtonEvent
	FsmToOrders        chan elevator.Elevator
	OrdersToPeers      chan elevator.NetworkElevator
	NetworkDisconnect  chan bool
	SkipCh             chan bool
}

// NewChannels initializes all channels
func NewChannels() *Channels {
	return &Channels{
		PeerUpdate:         make(chan peers.PeerUpdate),
		RemoteElevator:     make(chan msgTypes.ElevatorStateMsg),
		PeerTxEnable:       make(chan bool),
		LocalAssignedOrder: make(chan elevio.ButtonEvent),
		ButtonPress:        make(chan elevio.ButtonEvent),
		CompletedOrder:     make(chan elevio.ButtonEvent, 3),
		FsmToOrders:        make(chan elevator.Elevator),
		OrdersToPeers:      make(chan elevator.NetworkElevator),
		NetworkDisconnect:  make(chan bool),
		SkipCh:             make(chan bool),
	}
}

// setupBackupConnection establishes a UDP connection to the backup
func setupBackupConnection(backupPort string) (*net.UDPConn, error) {
	backupAddr, err := net.ResolveUDPAddr("udp", ":"+backupPort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve backup address: %v", err)
	}
	conn, err := net.DialUDP("udp", nil, backupAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to dial backup: %v", err)
	}
	return conn, nil
}

// syncWithBackup sends periodic heartbeats to the backup
func syncWithBackup(conn *net.UDPConn) {
	for {
		time.Sleep(500 * time.Millisecond)
		_, err := conn.Write([]byte("ping")) // Simple heartbeat
		if err != nil {
			log.Printf("Failed to send to backup: %v", err)
		}
	}
}
