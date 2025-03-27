package network

import (
	"Heis/pkg/elevator"
	"Heis/pkg/network/conn"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"time"
)


type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}


const interval = 15 * time.Millisecond
const timeout = 500 * time.Millisecond

// Transmits the elevator state and the id to all the other elevators on the elevatorState chanel.
// Keeps track of the conected elevators, and sends the elevatorstates on the elevatorStateCh to the order module, the ids are sent to main on the peerUpdate channel.
func Transmitter(port int, id string, transmitEnable <-chan bool, transmitterToRecivierSkipCh chan bool, ordersToPeersCH chan elevator.NetworkElevator) {
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



func Receiver(port int, selfId string, transmitterToRecivierSkipCh chan bool, peerUpdateCh chan<- PeerUpdate, elevatorStateCh chan<- ElevatorStateMsg) {
	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)


	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		var msg ElevatorStateMsg

		// If there was a send error, probably due to network disconnection recieve message from itself
		select {
		case <- transmitterToRecivierSkipCh:
			msg.Id = selfId
		default:
			err := json.Unmarshal(buf[:n], &msg)
			if err != nil {
				continue // Ignore invalid messages
			}			
		}

		id := msg.Id // Extract peer ID
		

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
			case elevatorStateCh <- msg:
			default:
			}
		}
	}
}