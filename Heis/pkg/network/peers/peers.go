package peers

import (
	"Heis/pkg/network/conn"
	"Heis/pkg/types"
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

func Transmitter(port int, id string, transmitEnable <-chan bool, elevatorState *types.Elevator) {
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-transmitEnable:
		case <-time.After(interval):
		}

		if enable {
			// Create elevator state message
			elevatorStateMsg := types.ElevatorStateMsg{
				Elevator: *elevatorState, // Pass by reference
				Id:       id,
			}

			// Convert to JSON
			data, err := json.Marshal(elevatorStateMsg)
			if err != nil {
				continue
			}

			// Send data
			conn.WriteTo(data, addr)
		}
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {
	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		var msg types.ElevatorStateMsg
		err := json.Unmarshal(buf[:n], &msg)
		if err != nil {
			continue // Ignore invalid messages
		}

		id := msg.Id // Extract peer ID

		// Adding new connection
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

		// Sending update (Only print IDs, not states)
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)

			peerUpdateCh <- p

			// Print connection IDs only
			fmt.Println("Connected peers:", p.Peers)
		}
	}
}
