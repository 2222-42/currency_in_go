package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"time"
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

func toInt(done <-chan interface{}, valueCh <-chan interface{}) <-chan int {
	intCh := make(chan int)
	go func() {
		defer close(intCh)
		for v := range valueCh {
			select {
			case <-done:
				return
			case intCh <- v.(int):
			}
		}
	}()
	return intCh
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

func isPrime(n int) bool {
	for i := 2; i < n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return n > 1
}

func primeFinder(done <-chan interface{}, intCh <-chan int) <-chan interface{} {

	primeCh := make(chan interface{})
	go func() {
		defer close(primeCh)
		for n := range intCh {
			if isPrime(n) {
				continue
			}
			select {
			case <-done:
				return
			case primeCh <- n:
			}
		}
	}()
	return primeCh
}

func fanIn(done <-chan interface{}, channels ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	multiplexedCh := make(chan interface{})

	multiplex := func(c <-chan interface{}) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case multiplexedCh <- i:
			}
		}
	}

	wg.Add(len(channels))
	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(multiplexedCh)
	}()

	return multiplexedCh
}

func main() {
	randInt := func() interface{} {
		return rand.Intn(50000000)
	}

	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	randIntStream := toInt(done, repeatFn(done, randInt))
	fmt.Println("Primes:")

	//primeStream := primeFinder(done, randIntStream)

	numFinders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", numFinders)

	finders := make([]<-chan interface{}, numFinders)
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream)
	}

	primeStream := fanIn(done, finders...)
	for prime := range take(done, primeStream, 10) {
		fmt.Printf("\t%d\n", prime)
	}

	fmt.Printf("Search took: %v", time.Since(start))
}
