package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	channel1 := make(chan int)
	channel2 := make(chan int)
	channel3 := make(chan int, 2)
	wg.Add(2)
	go func() {
		defer wg.Done()
		server(channel2, channel1, channel3)
	}()

	go func() {
		defer wg.Done()
		client(channel1, channel2, channel3)
	}()

	if <-channel3 == 1 {
		fmt.Println("client spazzed out")
	} else if <-channel3 == 2 {
		fmt.Println("server spazzed out")
	}
	wg.Wait()
}

func client(chanOut chan int, chanIn chan int, toMain chan int) {
	var done = false
	var attempts = 0
	for done == false {
		if attempts == 3 {
			fmt.Println("client timed out")
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
			} else {
				fmt.Println("server response not valid, resending")
				attempts++
			}
		case <-time.After(1 * time.Second):
			fmt.Println("server timed out, resending")
			attempts++
		}

	}

}

func server(chanOut chan int, chanIn chan int, toMain chan int) {
	var done = false
	var attempts = 0
	for done == false {
		var sSeq int = 300
		var cSeq int
		if attempts == 3 {
			fmt.Println("server timed out")
			toMain <- 2
			break
		}
		select {
		case clientresponse := <-chanIn:
			cSeq = clientresponse
			fmt.Println("SYNACK")
			chanOut <- cSeq + 1
			chanOut <- sSeq
		case <-time.After(1 * time.Second):
			fmt.Println("client timed out, resending")
			attempts++
			continue
		}
		select {
		case clientresponse := <-chanIn:
			if sSeq+1 == clientresponse {
				fmt.Println("SUCCESS")
				done = true
			} else {
				fmt.Println("FAILURE")
			}
		case <-time.After(1 * time.Second):
			fmt.Println("client timed out, resending")
			cSeq = 0
			attempts++
		}
	}

}
