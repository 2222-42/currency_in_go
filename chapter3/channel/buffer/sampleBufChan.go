package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	var stdoutBuff bytes.Buffer
	defer stdoutBuff.WriteTo(os.Stdout)

	intStream := make(chan int, 4)
	go func() {
		defer close(intStream)
		defer fmt.Fprintln(&stdoutBuff, "Producer Done.")
		for i := 0; i < 5; i++ {
			fmt.Fprintf(&stdoutBuff, "Sending: %d\n", i)
			intStream <- i
		}
	}()

	for integer := range intStream {
		fmt.Fprintf(&stdoutBuff, "Received %v.\n", integer)
	}

	chanOwner := func() <-chan int {
		resultCh := make(chan int, 5)
		go func() {
			defer close(resultCh)
			for i := 0; i <= 5; i++ {
				resultCh <- i
			}
		}()
		return resultCh
	}

	resultCh := chanOwner()
	for result := range resultCh {
		fmt.Printf("Received: %d\n", result)
	}
	fmt.Println("Done receiving!")
}
