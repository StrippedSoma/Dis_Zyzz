package main

import (
	"fmt"
	"time"
)

func main() {
	channel1 := make(chan int)
	channel2 := make(chan int)
	channel3 := make(chan int, 2)

	go func() {
		server(channel2, channel1, channel3)
	}()

	go func() {
		client(channel1, channel2, channel3)
	}()

	var x = <-channel3
	if x == 1 {
		fmt.Println("client timed out")
	} else if x == 2 {
		fmt.Println("server timed out")
	} else if x == 3 {
		fmt.Println("TCPIP SUCCESSFUL")
	}

}

func client(chanOut chan int, chanIn chan int, toMain chan int) {
	var done = false
	var attempts = 0
	for done == false {
		if attempts == 3 {
			fmt.Println("Client Error - 3 tries")
			toMain <- 1
			break
		}
		var cSeq int = 100
		var sSeq int = 0
		chanOut <- cSeq
		select {
		case serverresponse := <-chanIn:
			if cSeq+1 == serverresponse {
				cSeq = serverresponse
				sSeq = <-chanIn + 1
				chanOut <- sSeq
				done = true
				fmt.Println("Client recieved")
			} else {
				fmt.Println("Server response not valid, resending")
				attempts++
			}
		case <-time.After(1 * time.Second):
			fmt.Println("Server timed out, resending")
			attempts++
		}

	}
	fmt.Println("Client Loop Done")
}

func server(chanOut chan int, chanIn chan int, toMain chan int) {
	var done = false
	var sent = false
	var attempts = 0
	var cSeq int = <-chanIn

	for done == false {
		var sSeq int = 300
		if attempts == 3 {
			fmt.Println("Server Error - 3 tries")
			toMain <- 2
			break
		}
		if !sent {
			fmt.Println("Server responding")
			chanOut <- cSeq + 1
			chanOut <- sSeq
			sent = true
		}
		if sent {
			select {
			case clientresponse := <-chanIn:
				if sSeq+1 == clientresponse {
					fmt.Println("Server recieved - done")
					done = true
					toMain <- 3
				} else {
					fmt.Println("Client response not valid, resending")
					attempts++
					sent = false
				}
			case <-time.After(1 * time.Second):
				fmt.Println("Client timed out, resending")
				attempts++
				sent = false
			}
		}
	}
	fmt.Println("Server Loop Done")
}
