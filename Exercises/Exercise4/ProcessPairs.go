package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
	"os/exec"
)



func main () {

	fmt.Println("Backup")

	localadress_port := "30210"
	adress := ":" + localadress_port
	
	s, err := net.ResolveUDPAddr("udp", adress)
	if err != nil {
		return
	}
	
	//først lytte
	connection, err := net.ListenUDP("udp", s)

	if err != nil {
		fmt.Println("Error")
		return
	}


	buffer := make([]byte, 1024)
	counter := 0

	//read and update counter
	for {
		//udp detect timeout
		connection.SetReadDeadline(time.Now().Add(time.Millisecond * 1000))
		n, _, err := connection.ReadFromUDP(buffer)

		if err != nil {
			break
		}

		// message := ""
		// for _, element := range buffer {
		// 	if string(element) == 
		// 	message += string(element)
		// }

		// counter, err = strconv.Atoi(message)
		counter, err = strconv.Atoi(string(buffer[0:n]))
		if err != nil {
			fmt.Println("error")
		}
		// fmt.Println(message)
	}

	connection.Close()

	//primary

	connection, err = net.DialUDP("udp", nil, s)

	if err != nil {
		fmt.Println("Erro")
		return
	}

	//spawne ny av samme i annet vindu

	exec.Command("gnome-terminal", "--", "go", "run", "ProcessPairs.go").Run()
	fmt.Println("Takover as master")
	//så sende
	for {
		counter++
		data := []byte(strconv.Itoa(counter))
		time.Sleep(500 * time.Millisecond)
		_, err = connection.Write(data)
		fmt.Println(counter)
	}

}