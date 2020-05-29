package main

import (
	"fmt"
	"math/rand"
)

func doWork(
	done <-chan interface{},
) (<-chan interface{}, <-chan int) {
	heartbeat := make(chan interface{}, 1) // heartbeatの送信先チャネル。バッファ付きにして、常に最低1つの鼓動が送られることを保証する。
	workCh := make(chan int)

	go func() {
		defer close(heartbeat)
		defer close(workCh)

		for i := 0; i < 10; i++ {
			select {
			case heartbeat <- struct{}{}:
			default: //誰もハートビートを確認していない場合があるから
			}
			select {
			case <-done:
				return
			case workCh <- rand.Intn(10): // 送信や受信を行うときはいつでもハートびーとの鼓動に対する条件を含める必要がある

			}
		}
		// 入力を待つ間、あるいは、結果を送信するのを待っている間に、複数の鼓動を送信しているかもしれない
		// だから、select文をforループの中に置く必要があるんですね。
	}()

	return heartbeat, workCh
}

func main() {
	done := make(chan interface{})
	defer close(done)
	heartbeat, results := doWork(done)

	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				fmt.Println("worker goroutine is not healthy from heartbeat")
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				fmt.Println("worker goroutine is not healthy from result!")
				return
			}
			fmt.Printf("results %v \n", r)
			//case <-time.After(timeout):
			//	fmt.Println("worker goroutine is not healthy!")
			//	return
		}
	}
}
