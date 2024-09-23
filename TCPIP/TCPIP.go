package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	channel1 := make(chan int)
	channel2 := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		server(channel2, channel1)
	}()

	go func() {
		defer wg.Done()
		client(channel1, channel2)
	}()

	wg.Wait()
}

func client(toServer chan int, toClient chan int) {
	var cSeq int = 100
	var sSeq int = 0
	toServer <- cSeq

	if cSeq+1 == <-toClient {
		cSeq++
		sSeq = <-toClient + 1

		toServer <- sSeq
	} else {
		toServer <- 0
	}

}

func server(chanOut chan int, chanIn chan int) {
	var sSeq int = 300
	var cSeq int = 0
	cSeq = <-chanIn
	fmt.Println("SYNACK")
	chanOut <- cSeq + 1
	chanOut <- sSeq

	if sSeq+1 == <-chanIn {
		fmt.Println("SUCCESS")
	} else {
		fmt.Println("FAILURE")

	}
}
