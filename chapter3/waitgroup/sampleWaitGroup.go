package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("1st gotoutine sleeping...")
		time.Sleep(1)
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Println("2nd gotoutine sleeping...")
		time.Sleep(2)
	}()

	wg.Wait()
	fmt.Println("All goroutines complete.")
}
