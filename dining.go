package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

func main() {

	t := time.Now()

	var eatAmount int = 3

	chanIn1 := make(chan bool)
	chanIn2 := make(chan bool)
	chanIn3 := make(chan bool)
	chanIn4 := make(chan bool)
	chanIn5 := make(chan bool)

	chanOut1 := make(chan bool)
	chanOut2 := make(chan bool)
	chanOut3 := make(chan bool)
	chanOut4 := make(chan bool)
	chanOut5 := make(chan bool)
	chanOut6 := make(chan bool)
	chanOut7 := make(chan bool)
	chanOut8 := make(chan bool)
	chanOut9 := make(chan bool)
	chanOut10 := make(chan bool)

	go fork(chanIn1, chanOut1, chanOut2)
	go fork(chanIn2, chanOut3, chanOut4)
	go fork(chanIn3, chanOut5, chanOut6)
	go fork(chanIn4, chanOut7, chanOut8)
	go fork(chanIn5, chanOut9, chanOut10)

	wg.Add(5)

	go philosopher(chanIn1, chanIn2, chanOut1, chanOut4, eatAmount, t)
	go philosopher(chanIn2, chanIn3, chanOut3, chanOut6, eatAmount, t)
	go philosopher(chanIn3, chanIn4, chanOut5, chanOut8, eatAmount, t)
	go philosopher(chanIn4, chanIn5, chanOut7, chanOut10, eatAmount, t)
	go philosopher(chanIn5, chanIn1, chanOut9, chanOut2, eatAmount, t)

	wg.Wait()
	fmt.Println("Done")
}

func fork(listen chan bool, tellRight chan bool, tellLeft chan bool) {
	onTable := true
	side := true
	for {
		request := <-listen
		if request && onTable {
			tellRight <- true
			onTable = false
			side = true
		} else if !request && onTable {
			tellLeft <- true
			onTable = false
			side = false
		} else if request && !onTable {
			if side {
				onTable = true
			} else {
				tellRight <- false
			}
		} else if !request && !onTable {
			if !side {
				onTable = true
			} else {
				tellLeft <- false
			}
		}
	}
}

func philosopher(tellLeft chan bool, tellRight chan bool, listenLeft chan bool, listenRight chan bool, eatAmount int, t time.Time) {
	//fmt.Println("Philosopher is thinking")
	defer wg.Done()
	for eatAmount > 0 {
		tellLeft <- true
		if <-listenLeft == true {
			tellRight <- false
			if <-listenRight == true {
				//fmt.Println("Philosopher is eating", eatAmount)
				tellLeft <- true
				tellRight <- false
				time.Sleep(1 * time.Millisecond)
				eatAmount--
				//fmt.Println("Philosopher is thinking")
			} else {
				tellLeft <- true
				time.Sleep(1 * time.Millisecond)
			}
		}
	}
	fmt.Println("Philosopher is ------------------------------------------------- done", time.Since(t))
}
