package network

import (
	"Heis/pkg/config"
	"Heis/pkg/elevator"
	"Heis/pkg/network/conn"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

// defined in config
var (
	interval time.Duration
	timeout  time.Duration
)

// inits global variables from the config file
func init() {
	cfg, err := config.LoadConfig("config/elevator_params.json")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	interval = cfg.NetworkInterval * time.Millisecond
	timeout = cfg.NetworkTimeout * time.Millisecond
}

// Sends the information this elevator has for all the other elevators, to all the other elevators.
// If sending fails, probably due to network disconnection, transmitter sends on channel to reciver.
func Transmitter(port int, id string, 
	transmitEnable <-chan bool, transmitterToRecivierSkipCh chan bool, ordersToPeersCH chan elevator.NetworkElevator) {
	
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	// Make sure info is loaded before sending
	networkElevator := <-ordersToPeersCH

	enable := true
	for {
		// Should only send message once in an interval.
		// Also able to dissable and reanable sending.
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
		// Non-blocking check for updates on ordersToPeersCH
		select {
		case networkElevator = <-ordersToPeersCH:
		default:
		}

		if enable {
			// Create elevator state message
			elevatorStateMsg := ElevatorStateMsg{
				NetworkElevator: networkElevator,
				Id:              id,
			}

			// Convert to JSON
			data, err := json.Marshal(elevatorStateMsg)
			if err != nil {
				fmt.Println("send error:", err)
				continue
			}

			// Send data
			_, err = conn.WriteTo(data, addr)

			// If not able to send, forward a message to receiver to ensure correct handling of peers
			if err != nil {
				transmitterToRecivierSkipCh <- true
				continue
			}
		}
	}
}


// Handles incoming messages and keeps track of the connected elevators.
// Messages are sent on the elevatorStateCh when the reciving function is ready.
// A change in connected elevators results in a peerUpdate sent on the peerUpdateCh.
func Receiver(port int, selfId string, 
	transmitterToRecivierSkipCh chan bool, peerUpdateCh chan<- PeerUpdate, remoteElevatorUpdateCh chan<- ElevatorStateMsg) {
	
	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		var msg ElevatorStateMsg

		// If there was a send error, probably due to network disconnection recieve message from itself from transmitter
		select {
		case <-transmitterToRecivierSkipCh:
			msg.Id = selfId
		default:
			err := json.Unmarshal(buf[:n], &msg)
			if err != nil {
				// Ignore invalid messages
				continue 
			}
		}
		
		// Extract peer ID
		id := msg.Id 

		// Track peer presence
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}
			lastSeen[id] = time.Now()
		}

		// Removing dead connections
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Send peer update only if there was a change
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)

			peerUpdateCh <- p
		}

		// Forward the full elevator state to order module.
		// Non blocking to prevent back up of reciever.
		if (msg.Id != selfId) || (len(p.Peers) == 1) {
			select {
			case remoteElevatorUpdateCh <- msg:
			default:
			}
		}
	}
}
