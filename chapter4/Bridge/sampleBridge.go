package main

import "fmt"

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

func genVals() <-chan <-chan interface{} {
	chanCh := make(chan (<-chan interface{}))
	go func() {
		defer close(chanCh)
		for i := 0; i < 10; i++ {
			// バッファ付きチャネルにしないとデッドロックを起こす。
			ch := make(chan interface{}, 1)
			ch <- i
			close(ch)
			chanCh <- ch
		}
	}()
	return chanCh
}

func main() {
	for v := range bridge(nil, genVals()) {
		fmt.Printf("%v ", v)
	}
}
