package nettwork

import (
	"fmt"
	"net"
	"time"
)

func tcp_read(c net.Conn) {
	buffer := make([]byte, 1024)
	_, err := c.Read(buffer)

	if err != nil {
		fmt.Println(err)
	}

	message := ""
	for _, element := range buffer {
		message += string(element)
	}

	fmt.Println(message)
}

func read_loop(c net.Conn) {
	for {
		tcp_read(c)
	}
}

func tcp_send(c net.Conn, msg string) {
	time.Sleep(500 * time.Millisecond)
	paddedMessage := make([]byte, 1024)
	copy(paddedMessage, msg)
	c.Write(paddedMessage)
}

func send_loop(c net.Conn, msg string) {
	for {
		tcp_send(c, msg)
	}
}

func Client(adress string) {
	Connect := adress

	c, err := net.Dial("tcp", Connect)

	if err != nil {
		fmt.Println(err)
	}

	v := make(chan int)

	go read_loop(c)
	go send_loop(c, "Hei")

	<-v

	defer c.Close()
}

func Server(portnumber string) {
	c0, err := net.Dial("tcp", "10.100.23.204:33546")
	tcp_send(c0, "Connect to: "+"10.100.23.18"+":"+portnumber)
	//c.Write([]byte("Connect to: "+ "127.0.0.1/8" +":"+ "20008" +"\0"))
	PORT := ":" + portnumber

	l, err := net.Listen("tcp", PORT) //kan også bruke net.ListenTCP

	if err != nil {
		fmt.Println(err)
	}

	c, err := l.Accept()
	fmt.Println("her")

	if err != nil {
		fmt.Println(err)
	}

	v := make(chan int)
	go read_loop(c)
	go send_loop(c, "hasd")

	defer l.Close()
	<-v

}
