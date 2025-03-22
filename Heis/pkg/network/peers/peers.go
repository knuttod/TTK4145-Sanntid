package peers

import (
	"Heis/pkg/elevator"
	"Heis/pkg/msgTypes"
	"Heis/pkg/network/conn"

	// "Heis/pkg/orders"
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
func Transmitter(port int, id string, transmitEnable <-chan bool, nettworkDisconnectCh chan bool, ordersToPeersCH chan elevator.NetworkElevator) {
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	// Make sure info is loaded before sending
	networkElevator := <-ordersToPeersCH

	fmt.Println("start sending")
	iter := 0

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}
			// Non-blocking check for updates on ordersToPeersCH
			select {
			case networkElevator = <-ordersToPeersCH:
				// Update the networkElevator data
			default:
				// No update available, proceed with sending
			}

			// fmt.Println("send")

			if enable {
				iter++
				// Create elevator state message
				elevatorStateMsg := msgTypes.ElevatorStateMsg{
					NetworkElevator: networkElevator,
					Id:              id,
					Iter:            iter,
				}

				// Convert to JSON
				data, err := json.Marshal(elevatorStateMsg)
				if err != nil {
					fmt.Println("send error:", err)
					continue
				}

				// Send data
				_, err = conn.WriteTo(data, addr)
				if err != nil {
					fmt.Println("Send error:", err)
					fmt.Println("her")
					nettworkDisconnectCh <- true
					continue
				}
				if id == "heis1" {
					// fmt.Println("send iter", iter)
				}
				// fmt.Println("Data sent successfully")
			// }
		}
	}
}



func Receiver(port int, selfId string, peerUpdateCh chan<- PeerUpdate, elevatorStateCh chan<- msgTypes.ElevatorStateMsg) {
	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)


	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		var msg msgTypes.ElevatorStateMsg

		//Jalla løsning
		// if (len(p.Peers) == 1) && (p.Peers[0] == selfId) || (len(p.Peers) == 0) {
		// 	msg.Id = selfId
		// 	select {
		// 	case elevatorStateCh <- msg:
		// 		// Successfully sent to channel
		// 	default:
		// 		// Channel is full, skipping send
		// 	}
		// } 

		err := json.Unmarshal(buf[:n], &msg)
		if err != nil {
			// fmt.Println(err)
			continue // Ignore invalid messages
		}
		// fmt.Println("comp")

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

		// Forward the full elevator state to order module
		if (msg.Id != selfId) || (len(p.Peers) == 1) {
			select {
			case elevatorStateCh <- msg:
				// Successfully sent to channel
			default:
				// Channel is full, skipping send
			}
		}
	}
}