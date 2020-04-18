package main

import (
	"fmt"
	"sync"
	"time"
)

func main() {
	c := sync.NewCond(&sync.Mutex{})

	queue := make([]interface{}, 0, 10)
	removeFromQueue := func(delay time.Duration) {
		time.Sleep(delay)
		c.L.Lock()
		queue = queue[1:]
		fmt.Println("Removed from queue")
		// Unlockなのに`defer`していない箇所がある理由は？Signalを送る前にUnlockしたいからか？
		// c.L.Unlock()
		// defer しても普通に動くが？
		defer c.L.Unlock()
		c.Signal()
	}

	for i := 0; i < 10; i++ {
		c.L.Lock()
		// `for`で囲む理由は、シグナルが来るまでWaitによって待つが、しかしそのシグナルが求めているシグナルかは不明だから。
		for len(queue) == 2 {
			c.Wait()
		}
		fmt.Println("Adding to queue")
		queue = append(queue, struct{}{})
		go removeFromQueue(1 * time.Second)
		c.L.Unlock()
	}
}
