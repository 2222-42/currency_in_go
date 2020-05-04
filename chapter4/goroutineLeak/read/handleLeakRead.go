package read

import (
	"fmt"
	"time"
)

func main() {
	doWork := func(
		done <-chan interface{},
		strings <-chan string,
	) <-chan interface{} { // 慣例としてdoneチャネルは第一引数に
		completed := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(completed)
			//for s := range strings {
			//	fmt.Println(s)
			//}
			for {
				select {
				case s := <-strings:
					fmt.Println(s)
				case <-done:
					return
				}
			}
		}()
		return completed
	}
	done := make(chan interface{})
	terminated := doWork(done, nil)
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Canceling doWork goroutine.")
		close(done)
	}()

	<-terminated // こでdoWorkから生成されたゴルーチンがメインゴルーチンにつながる

	fmt.Println("Done.")
}
