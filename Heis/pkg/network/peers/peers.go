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

// Får en delay i meldingene. Den som sender motar sin egen melding en del før den som ikke sender
// er dette problem med udp?
const interval = 15 * time.Millisecond
// const interval = 220 * time.Millisecond
const timeout = 500 * time.Millisecond

// Transmits the elevator state and the id to all the other elevators on the elevatorState chanel.
// func Transmitter(port int, id string, transmitEnable <-chan bool, ordersToPeersCH chan elevator.NetworkElevator) {
// 	conn := conn.DialBroadcastUDP(port)
// 	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

// 	//Make sure info is loaded before sending
// 	networkElevator := <- ordersToPeersCH

// 	fmt.Println("start sending")

// 	enable := true
// 	send := true
// 	for {
// 		select {
// 		case enable = <-transmitEnable:
// 			send = true
// 		case <-time.After(interval):
// 			fmt.Println("nu")
// 			send = true
// 		// case networkElevator = <- ordersToPeersCH:
// 		// 	send = false
// 		// 	fmt.Println("updat")
// 		}

// 		if enable && send{
// 			// Create elevator state message
// 			elevatorStateMsg := msgTypes.ElevatorStateMsg{
// 				NetworkElevator: networkElevator,
// 				Id:       id,
// 			}
// 			fmt.Println("Send")

// 			// Convert to JSON
// 			data, err := json.Marshal(elevatorStateMsg)
// 			if err != nil {
// 				fmt.Println("send error")
// 				continue
// 			}

// 			// Send data
// 			conn.WriteTo(data, addr)
// 		}
// 	}
// }

func Transmitter(port int, id string, transmitEnable <-chan bool, ordersToPeersCH chan elevator.NetworkElevator) {
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

// // Keeps track of the conected elevators, and sends the elevatorstates on the elevatorStateCh to the order module, the ids are sent to main on the peerUpdate channel.
// func Receiver(port int, peerUpdateCh chan<- PeerUpdate, elevatorStateCh chan<- msgTypes.ElevatorStateMsg) {
// 	var buf [1024]byte
// 	var p PeerUpdate
// 	lastSeen := make(map[string]time.Time)

// 	conn := conn.DialBroadcastUDP(port)

// 	for {
// 		updated := false

// 		conn.SetReadDeadline(time.Now().Add(interval))
// 		n, _, _ := conn.ReadFrom(buf[0:])

// 		var msg msgTypes.ElevatorStateMsg
// 		err := json.Unmarshal(buf[:n], &msg)
// 		if err != nil {
// 			// fmt.Println(err)
// 			continue // Ignore invalid messages
// 		}

// 		id := msg.Id // Extract peer ID

// 		// Track peer presence
// 		p.New = ""
// 		if id != "" {
// 			if _, idExists := lastSeen[id]; !idExists {
// 				p.New = id
// 				updated = true
// 			}
// 			lastSeen[id] = time.Now()
// 		}

// 		// Removing dead connections
// 		p.Lost = make([]string, 0)
// 		for k, v := range lastSeen {
// 			if time.Now().Sub(v) > timeout {
// 				updated = true
// 				p.Lost = append(p.Lost, k)
// 				delete(lastSeen, k)
// 			}
// 		}

// 		// Send peer update only if there was a change
// 		if updated {
// 			p.Peers = make([]string, 0, len(lastSeen))

// 			for k := range lastSeen {
// 				p.Peers = append(p.Peers, k)
// 			}

// 			sort.Strings(p.Peers)
// 			sort.Strings(p.Lost)
// 			fmt.Println("beg")
// 			peerUpdateCh <- p
// 			fmt.Println("late")

// 			fmt.Println("Connected peers:", p.Peers)
// 		}

// 		// Forward the full elevator state to order module
// 		elevatorStateCh <- msg
// 	}
// }

func Receiver(port int, selfId string, peerUpdateCh chan<- PeerUpdate, elevatorStateCh chan<- msgTypes.ElevatorStateMsg) {
	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	//temp
	updateFlag := false
	var updatePeers PeerUpdate

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		// startTime := time.Now() // Capture start time

		var msg msgTypes.ElevatorStateMsg
		err := json.Unmarshal(buf[:n], &msg)
		if err != nil {
			// fmt.Println("err")
			continue // Ignore invalid messages
		}

		// duration := time.Since(startTime) // Calculate duration
		// fmt.Printf("json.Unmarshal took %s\n", duration)

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
		// if updated {
		// 	p.Peers = make([]string, 0, len(lastSeen))

		// 	for k := range lastSeen {
		// 		p.Peers = append(p.Peers, k)
		// 	}

		// 	sort.Strings(p.Peers)
		// 	sort.Strings(p.Lost)

		// 	select {
		// 	case peerUpdateCh <- p:
		// 		// Successfully sent to channel
		// 	default:
		// 		// Channel is full, skipping send
		// 		fmt.Println("peerUpdateCh channel is full, skipping send")
		// 	}

		// 	fmt.Println("Connected peers:", p.Peers)
		// }

		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)

			updatePeers = p
			updateFlag = true
		}

		if updateFlag {
			fmt.Println("change in peers")

			select {
			case peerUpdateCh <- updatePeers:
				// Successfully sent to channel
				fmt.Println("Connected peers:", updatePeers.Peers)
				updateFlag = false
			// default:
				// Channel is full, skipping send
				// fmt.Println("peerUpdateCh channel is full, skipping send")
			}

			
		}

		// Forward the full elevator state to order module
		//legge til buffer for forskjellige heiser
		if msg.Id != selfId {
			select {
			case elevatorStateCh <- msg:
				// Successfully sent to channel
			default:
				// Channel is full, skipping send
				// fmt.Println("elevatorStateCh channel is full, skipping send")
			}
		}
		if msg.Id == "heis1" {

			// fmt.Println("Id:", msg.Id, "Iter:", msg.Iter)
		}
		

		// fmt.Println("Id:", msg.Id)
	}
}

// func Receiver(port int, peerUpdateCh chan<- PeerUpdate, elevatorStateCh chan<- msgTypes.ElevatorStateMsg) {
//     var buf [1024]byte
//     var p PeerUpdate
//     lastSeen := make(map[string]time.Time)

//     conn := conn.DialBroadcastUDP(port)

//     for {
//         updated := false

//         conn.SetReadDeadline(time.Now().Add(interval))
//         n, _, err := conn.ReadFrom(buf[0:])
//         if err != nil {
//             if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
// 				// fmt.Println("Timeout")
//                 // Timeout is expected, continue to the next iteration
//                 continue
//             }
//             fmt.Println("Read error:", err)
//             continue
//         }

//         if n == 0 {
//             // No data read, continue to the next iteration
//             continue
//         }

//         var msg msgTypes.ElevatorStateMsg
//         err = json.Unmarshal(buf[:n], &msg)
//         if err != nil {
//             fmt.Println("JSON unmarshal error:", err)
//             continue // Ignore invalid messages
//         }

//         id := msg.Id // Extract peer ID

//         // Forward the full elevator state to order module
//         elevatorStateCh <- msg

//         // Track peer presence
//         p.New = ""
//         if id != "" {
//             if _, idExists := lastSeen[id]; !idExists {
//                 p.New = id
//                 updated = true
//             }
//             lastSeen[id] = time.Now()
//         }

//         // Removing dead connections
//         p.Lost = make([]string, 0)
//         for k, v := range lastSeen {
//             if time.Now().Sub(v) > timeout {
//                 updated = true
//                 p.Lost = append(p.Lost, k)
//                 delete(lastSeen, k)
//             }
//         }

//         // Send peer update only if there was a change
//         if updated {
//             p.Peers = make([]string, 0, len(lastSeen))

//             for k := range lastSeen {
//                 p.Peers = append(p.Peers, k)
//             }

//             sort.Strings(p.Peers)
//             sort.Strings(p.Lost)

//             peerUpdateCh <- p

//             fmt.Println("Connected peers:", p.Peers)
//         }
//     }
// }
