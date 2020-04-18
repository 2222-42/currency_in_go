package main

import (
	"fmt"
	"sync"
)

func main() {
	var count int

	increment := func() {
		count++
	}

	decrement := func() {
		count--
	}

	var onceA, onceB sync.Once

	var increments sync.WaitGroup
	increments.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer increments.Done()
			onceA.Do(increment)
		}()
	}
	onceB.Do(decrement)

	increments.Wait()
	fmt.Printf("Count is %d", count)
}
