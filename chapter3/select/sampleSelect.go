package main

import (
	"fmt"
	"time"
)

func main() {
	//start := time.Now()
	done := make(chan interface{})
	go func() {
		time.Sleep(5 * time.Second)
		close(done)
	}()
	//var c1, c2 <-chan interface{}
	//var c3 chan<- interface{}

	workCounter := 0
loop:
	for {
		select {
		case <-done:
			break loop
		default: // このdefault節がないとselect文でブロックされ続ける。
		}
		workCounter++
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("Achived %v cycles of work before signalled to sto.\n", workCounter)
	//select {
	//case <-c:
	//	fmt.Printf("Unblocked %v later.\n", time.Since(start))
	//case <- c1:
	//	println("case1")
	//case <- c2:
	//	println("case2")
	//case c3<- struct {}{}:
	//	println("case3")
	//case <-time.After(1 * time.Second):
	//	fmt.Println("Timed out.")
	//default:
	//	fmt.Printf("In default after %v\n\n", time.Since(start))
	//}
}
