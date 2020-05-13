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

func repeat(done <-chan interface{}, values ...interface{}) <-chan interface{} {
	valueCh := make(chan interface{})
	go func() {
		defer close(valueCh)
		for {
			for _, v := range values {
				select {
				case <-done:
					return
				case valueCh <- v:
				}
			}
		}
	}()

	return valueCh
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

func tee(done <-chan interface{}, in <-chan interface{}) (_, _ <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})
	go func() {
		defer close(out1)
		defer close(out2)

		for val := range orDone(done, in) {
			var out1, out2 = out1, out2 // コピー変数としてローカル変数を用意
			for i := 0; i < 2; i++ {
				select {
				case out1 <- val:
					out1 = nil
				case out2 <- val:
					out2 = nil
				}
			}
		}
	}()
	return out1, out2
}

func main() {
	done := make(chan interface{})
	defer close(done)

	out1, out2 := tee(done, take(done, repeat(done, 1, 2), 4))

	for val1 := range out1 {
		fmt.Printf("out1: %v, out2: %v\n", val1, <-out2)
	}
}
