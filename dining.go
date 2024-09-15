package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	t := time.Now()
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
			}
		} else {
			channel <- false
		}
	}
}

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
				time.Sleep(1 * time.Millisecond)
				eatAmount--
				//fmt.Println("Philosopher is thinking")
			} else {
				right <- false
				time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			}
		}
	}
	fmt.Println("Philosopher is ------------------------------------------------- done", time.Since(t))
}
