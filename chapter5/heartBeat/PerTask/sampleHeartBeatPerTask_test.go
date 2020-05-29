package main

import (
	"testing"
	"time"
)

func DoWork(done <-chan interface{}, nums []int) (<-chan interface{}, <-chan int) {
	heartbeat := make(chan interface{}, 1) // heartbeatの送信先チャネル。バッファ付きにして、常に最低1つの鼓動が送られることを保証する。
	intCh := make(chan int)

	go func() {
		defer close(heartbeat)
		defer close(intCh)

		time.Sleep(2 * time.Second)

		for _, n := range nums {
			select {
			case heartbeat <- struct{}{}:
			default: //誰もハートビートを確認していない場合があるから
			}
			select {
			case <-done:
				return
			case intCh <- n: // 送信や受信を行うときはいつでもハートびーとの鼓動に対する条件を含める必要がある
			}
		}
		// 入力を待つ間、あるいは、結果を送信するのを待っている間に、複数の鼓動を送信しているかもしれない
		// だから、select文をforループの中に置く必要があるんですね。
	}()

	return heartbeat, intCh
}

func TestDoWork_GeneratesAllNumbers(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	intSlice := []int{0, 1, 2, 3, 5}
	heartbeat, results := DoWork(done, intSlice)

	<-heartbeat // ゴルーチンが繰り返しを始めるというシグナルを送るのを待つ

	//for i, expected := range intSlice {
	//	select {
	//	case r := <-results:
	//		if r != expected{
	//			t.Errorf(
	//				"index %v: expected %v, but received %v,",
	//				i,
	//				expected,
	//				r,
	//				)
	//		}
	//	case <-time.After(1 * time.Second):
	//		t.Fatal("test timed out")
	//	}
	//}

	i := 0
	for r := range results {
		if expected := intSlice[i]; r != expected {
			t.Errorf(
				"index %v: expected %v, but received %v,",
				i,
				expected,
				r,
			)
		}
		i++
	}
}
