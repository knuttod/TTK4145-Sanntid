package nettwork

import (
	"fmt"
	"net"
	"time"
)

// Receiver/server

func Receiver(portnumber string) {

	//portnumber := "30000"
	PORT := ":" + portnumber

	s, err := net.ResolveUDPAddr("udp", PORT)

	if err != nil {
		fmt.Println("Erro")
		return
	}

	connection, err := net.ListenUDP("udp", s)

	if err != nil {
		fmt.Println("Erro")
		return
	}

	defer connection.Close()

	buffer := make([]byte, 1024)
	for {
		nBytesReceived, fromWho, err := connection.ReadFromUDP(buffer)

		if err != nil {
			fmt.Println("Erro")
			return
		}

		message := ""
		for _, element := range buffer {
			message += string(element)
		}

		fmt.Println(message)
		fmt.Println(nBytesReceived)
		fmt.Println(fromWho)
		//return
	}
}

func Sender(adress string) {

	//Connect := "10.100.23.204:20008"
	Connect := adress

	s, err := net.ResolveUDPAddr("udp", Connect)
	if err != nil {
		fmt.Println("Erro")
		return
	}
	c, err := net.DialUDP("udp", nil, s)
	if err != nil {
		fmt.Println("Erro")
		return
	}

	defer c.Close()

	for {
		data := []byte("HELLO\n")
		time.Sleep(500 * time.Millisecond)
		_, err = c.Write((data))
		time.Sleep(500 * time.Millisecond)

		fmt.Println("Test")
	}

}
