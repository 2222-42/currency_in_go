package main

import "fmt"

func generater(done <-chan interface{}, intergers ...int) <-chan int {
	intCh := make(chan int, len(intergers))

	go func() {
		defer close(intCh)
		for _, i := range intergers {
			select {
			case <-done:
				return
			case intCh <- i:
			}
		}
	}()
	return intCh
}

func multiply(
	done <-chan interface{},
	intCh <-chan int,
	multiplier int,
) <-chan int {
	multipliedCh := make(chan int)
	go func() {
		defer close(multipliedCh)
		for i := range intCh {
			select {
			case <-done:
				return
			case multipliedCh <- i * multiplier:
			}
		}
	}()
	return multipliedCh
}

func add(
	done <-chan interface{},
	intCh <-chan int,
	additive int,
) <-chan int {
	addedCh := make(chan int)
	go func() {
		defer close(addedCh)
		for i := range intCh {
			select {
			case <-done:
				return
			case addedCh <- i + additive:
			}
		}
	}()
	return addedCh
}

func main() {
	done := make(chan interface{})
	defer close(done)

	intCh := generater(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, intCh, 2), 1), 2)

	for v := range pipeline {
		fmt.Println(v)
	}
}
