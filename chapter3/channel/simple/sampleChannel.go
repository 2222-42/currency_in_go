package main

import (
	"fmt"
	"sync"
)

func main() {
	//var receiveChan <-chan interface{}
	//var sendChan chan<- interface{}
	//dataStream := make(chan interface{})
	//
	//receiveChan = dataStream
	//sendChan = dataStream

	stringCh := make(chan string)
	go func() {
		//if true {
		//	return
		//}
		stringCh <- "Hello channels!"
	}()
	salutation, ok := <-stringCh
	fmt.Printf("(%v): %v\n", ok, salutation)
	//// The following should fail as deadlock because of channel is empty and is open
	//salutation2, ok := <-stringCh
	//fmt.Printf("(%v): %v\n", ok, salutation2)

	//// The following codes are failed when compiling
	//readChan := make(<-chan interface{})
	//writeChan := make(chan<- interface{})
	//<- writeChan
	//readChan <- struct {}{}

	intCh := make(chan int)
	go func() {
		// If not close, goroutine became block and it produces deadlock
		defer close(intCh)
		for i := 1; i <= 5; i++ {
			intCh <- i
		}
	}()
	for integer := range intCh {
		fmt.Printf("%v ", integer)
	}

	begin := make(chan interface{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-begin
			fmt.Printf("%v has bugn\n", i)
		}(i)
	}

	fmt.Println("unblocking goroutines..")
	close(begin)
	wg.Wait()
}
