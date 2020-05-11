package main

import (
	"fmt"
	"math/rand"
)

func repeatFn(done <-chan interface{}, fn func() interface{}) <-chan interface{} {
	valueCh := make(chan interface{})
	go func() {
		defer close(valueCh)
		for {
			select {
			case <-done:
				return
			case valueCh <- fn():
			}
		}
	}()

	return valueCh
}

func take(done <-chan interface{}, valueCh <-chan interface{}, num int) <-chan interface{} {
	takeCh := make(chan interface{})
	go func() {
		defer close(takeCh)
		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case takeCh <- <-valueCh:
			}
		}
	}()
	return takeCh
}

func main() {
	done := make(chan interface{})
	defer close(done)

	randInt := func() interface{} { return rand.Int() }

	for num := range take(done, repeatFn(done, randInt), 10) {
		fmt.Println(num)
	}
}
