package main

import (
	//"fmt"
	"src/nettwork"
)

func main() {

	//a := make(chan int)

	// go nettwork.Receiver("20008")
	// go nettwork.Sender("10.100.23.204:20008")

	// <-a

	// fmt.Println("done")
	//nettwork.Sender("asdf")

	//Client("10.100.23.204:33546")

	nettwork.Server("20008")
}
