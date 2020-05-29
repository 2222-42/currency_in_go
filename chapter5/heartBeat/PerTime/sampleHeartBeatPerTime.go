package main

import (
	"fmt"
	"time"
)

func doWork(
	done <-chan interface{},
	pulseInterval time.Duration,
) (<-chan interface{}, <-chan time.Time) {
	heartbeat := make(chan interface{}) // heartbeatの送信先チャネル
	results := make(chan time.Time)

	go func() {
		defer close(heartbeat)
		defer close(results)

		pulse := time.Tick(pulseInterval)
		workGen := time.Tick(2 * pulseInterval) // 仕事が入ってくる様子のシミュレートに使う

		sendPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default: //誰もハートビートを確認していない場合があるから
			}
		}

		sendResult := func(r time.Time) {
			for {
				select {
				case <-done: // 失敗するサンプルケースではこのcaseの部分がないがそれは本質的か？
					return
				case <-pulse:
					sendPulse()
				case results <- r:
					return
				}
			}
		}
		for i := 0; i < 2; i++ {
			select {
			case <-done:
				return
			case <-pulse: // 送信や受信を行うときはいつでもハートびーとの鼓動に対する条件を含める必要がある
				sendPulse()
			case r := <-workGen:
				sendResult(r)
			}
		}
		// 入力を待つ間、あるいは、結果を送信するのを待っている間に、複数の鼓動を送信しているかもしれない
		// だから、select文をforループの中に置く必要があるんですね。
	}()

	return heartbeat, results
}

func main() {
	done := make(chan interface{})
	time.AfterFunc(10*time.Second, func() {
		close(done)
	})

	const timeout = 2 * time.Second
	heartbeat, results := doWork(done, timeout/2)

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
			fmt.Printf("results %v \n", r.Second())
		case <-time.After(timeout):
			fmt.Println("worker goroutine is not healthy!")
			return
		}
	}
}
