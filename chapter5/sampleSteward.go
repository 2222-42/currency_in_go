package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

// 監視と再起動ができるゴルーチンのシグネチャを定義
type startGoroutineFn func(
	done <-chan interface{},
	pulseInterval time.Duration,
) (heartbeat <-chan interface{})

func or(channels ...<-chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}
	orDone := make(chan interface{})
	go func() {
		defer close(orDone)
		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
			case <-or(append(channels[3:], orDone)...):
			}
		}
	}()
	return orDone
}

func newSteward(
	timeout time.Duration,
	startGoroutine startGoroutineFn,
) startGoroutineFn { // 監視するゴルーチンのためのtimeoutと監視するゴルーチンを起動するためのstartGoroutineFnを取る。帰り値より管理人も監視可能である。
	return func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
		heartbeat := make(chan interface{})
		go func() {
			defer close(heartbeat)

			var wardDone chan interface{}
			var wardHeartBeat <-chan interface{}
			startWard := func() { // 監視しているゴルーチンを起動するための一環した方法としてクロージャーを定義
				wardDone = make(chan interface{})                             // 中庭に渡すための新しいチャネル
				wardHeartBeat = startGoroutine(or(wardDone, done), timeout/2) // 監視対象のゴルーチンの起動
			}
			startWard()
			pulse := time.Tick(pulseInterval)

		monitorLoop:
			for {
				timeoutSignal := time.After(timeout)

				for {
					select {
					case <-pulse:
						select {
						case heartbeat <- struct{}{}:
						default:
						}
					case <-wardHeartBeat:
						continue monitorLoop
					case <-timeoutSignal:
						log.Println("steward: ward unhealthy; restarting")
						close(wardDone)
						startWard()
						continue monitorLoop
					case <-done:
						return
					}
				}
			}
		}()

		return heartbeat
	}
}

func orDone(done, c <-chan interface{}) <-chan interface{} {
	valCh := make(chan interface{})
	go func() {
		defer close(valCh)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case valCh <- v:
				case <-done:
				}

			}
		}
	}()
	return valCh
}

func bridge(
	done <-chan interface{},
	chanCh <-chan <-chan interface{},
) <-chan interface{} {
	valCh := make(chan interface{})
	go func() {
		defer close(valCh)
		for {
			var ch <-chan interface{}
			select {
			case maybeCh, ok := <-chanCh:
				if ok == false {
					return
				}
				ch = maybeCh
			case <-done:
				return
			}
			for val := range orDone(done, ch) {
				select {
				case valCh <- val:
				case <-done:
				}
			}

		}
	}()
	return valCh
}

func doWrokFn(done <-chan interface{}, intList ...int) (startGoroutineFn, <-chan interface{}) {
	intChCh := make(chan (<-chan interface{}))
	intCh := bridge(done, intChCh) // 中庭のインスタンスを複数起動する可能性があるからbridgeチャネルを使う。単一の妨げられないチャネルをdoWorkの消費者に渡す手助けをする。
	doWork := func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) <-chan interface{} { // 管理人に監視されるクロージャーを作成。
		intCh := make(chan interface{})
		heartbeat := make(chan interface{})
		go func() {
			defer close(intCh)
			select {
			case intChCh <- intCh: // bridgeチャンネルに、やり取りに使う新しいチャネルを知らせる
			case <-done:
				return
			}

			pulse := time.Tick(pulseInterval)

			for {
			valueloop:
				//for _, intVal := range intList {
				for {
					intVal := intList[0]
					intList = intList[1:]
					if intVal < 0 {
						log.Printf("negative value: %v\n", intVal) // 不健全な中庭をシミュレート
						return                                     // ここでreturnするからheartbeat に送られず、stewardで再起動される
					}

					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct{}{}:
								continue valueloop
							default:
							}
						case intCh <- intVal:
							continue valueloop
						case <-done:
							return
						}
					}
				}
			}
		}()
		return heartbeat
	}

	return doWork, intCh
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

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	//doWork := func(done <-chan interface{}, _ time.Duration) <-chan interface{} {
	//	log.Println("ward: Hello, I'm irresponsible!")
	//	go func() {
	//		<-done // キャンセルされるのを待ち続け、何もしない
	//		log.Println("ward: I am halting.")
	//	}()
	//	return nil
	//}
	done := make(chan interface{})
	defer close(done)
	doWork, intCh := doWrokFn(done, 1, 2, -1, 3, 4, 5)
	doWorkWithSteward := newSteward(1*time.Millisecond, doWork)
	doWorkWithSteward(done, 1*time.Hour)

	for intVal := range take(done, intCh, 6) {
		fmt.Printf("Received: %v\n", intVal)
	}

	//time.AfterFunc(9*time.Second, func() {
	//	log.Println("main: halting steward and ward.")
	//	close(done)
	//})
	//
	//for range doWorkWithSteward(done, 4*time.Second) {
	//
	//}
	log.Println("Done")
}
