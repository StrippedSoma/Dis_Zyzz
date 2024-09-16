package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	t := time.Now()
	var wg sync.WaitGroup
	var eatAmount int = 3
	chan1 := make(chan bool)
	chan2 := make(chan bool)
	chan3 := make(chan bool)
	chan4 := make(chan bool)
	chan5 := make(chan bool)

	go fork(chan1)
	go fork(chan2)
	go fork(chan3)
	go fork(chan4)
	go fork(chan5)

<<<<<<< Updated upstream
	go philosopher(chan1, chan2, eatAmount, t)
	go philosopher(chan2, chan3, eatAmount, t)
	go philosopher(chan3, chan4, eatAmount, t)
	go philosopher(chan4, chan5, eatAmount, t)
	go philosopher(chan5, chan1, eatAmount, t)

	time.Sleep(5 * time.Second)
	fmt.Println("Done")
}

func fork(channel chan bool) {
	taken := false
	for {
		request := <-channel
		change := request != taken
		taken = true
		if change {
			if request {
				channel <- true
			} else {
				taken = false
=======
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
>>>>>>> Stashed changes
			}
		} else {
			channel <- false
		}
	}
}

<<<<<<< Updated upstream
func philosopher(left chan bool, right chan bool, eatAmount int, t time.Time) {
	//fmt.Println("Philosopher is thinking")
	for eatAmount > 0 {
		right <- true
		if <-right == true {
			left <- true
			if <-left == true {
				//fmt.Println("Philosopher is eating", eatAmount)
				left <- false
				right <- false
=======
func philosopher(tellLeft chan bool, tellRight chan bool, listenLeft chan bool, listenRight chan bool, eatAmount int, t time.Time) {
	//fmt.Println("Philosopher is thinking")
	for eatAmount > 0 {
		tellLeft <- true
		if <-listenLeft == true {
			tellRight <- false
			if <-listenRight == true {
				//fmt.Println("Philosopher is eating", eatAmount)
				tellLeft <- true
				tellRight <- false
>>>>>>> Stashed changes
				time.Sleep(1 * time.Millisecond)
				eatAmount--
				//fmt.Println("Philosopher is thinking")
			} else {
<<<<<<< Updated upstream
				right <- false
=======
				tellLeft <- true
>>>>>>> Stashed changes
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}
	}
	fmt.Println("Philosopher is ------------------------------------------------- done", time.Since(t))
}
