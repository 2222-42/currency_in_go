package main

import (
	"fmt"
	"sync"
)

func main() {
	type Button struct {
		Clicked *sync.Cond
	}

	button := Button{Clicked: sync.NewCond(&sync.Mutex{})}

	subscribe := func(c *sync.Cond, fn func()) {
		var goroutineRunnning sync.WaitGroup
		goroutineRunnning.Add(1)
		go func() {
			goroutineRunnning.Done()
			c.L.Lock()
			defer c.L.Unlock()
			c.Wait()
			fn()
		}()
		goroutineRunnning.Wait()
	}

	var clickRegistered sync.WaitGroup
	clickRegistered.Add(3)
	subscribe(button.Clicked, func() {
		fmt.Println("Maximizing windows.")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Displaying annoying dialog box!")
		clickRegistered.Done()
	})
	subscribe(button.Clicked, func() {
		fmt.Println("Mouse clicked.")
		clickRegistered.Done()
	})
	//	これがないと"fatal error: all goroutines are asleep - deadlock!"となる。
	button.Clicked.Broadcast()
	clickRegistered.Wait()

}
