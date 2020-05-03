package main

import (
	"bytes"
	"fmt"
	"os"
)

func main() {
	var stdoutBuff bytes.Buffer
	defer stdoutBuff.WriteTo(os.Stdout)

	intStream := make(chan int, 4)
	go func() {
		defer close(intStream)
		defer fmt.Fprintln(&stdoutBuff, "Producer Done.")
		for i := 0; i < 5; i++ {
			fmt.Fprintf(&stdoutBuff, "Sending: %d\n", i)
			intStream <- i
		}
	}()

	for integer := range intStream {
		fmt.Fprintf(&stdoutBuff, "Received %v.\n", integer)
	}

	chanOwner := func() <-chan int {
		resultCh := make(chan int, 5) // チャネルをレキシカルスコープ内で初期化。書き込みできるスコープの制限、権限の拘束
		go func() {
			defer close(resultCh)
			for i := 0; i <= 5; i++ {
				resultCh <- i
			}
		}()
		return resultCh
	}

	consumer := func(resultCh <-chan int) { // チャネルへの読み込み権限を受け取り、読み込み以外の何もしないように、拘束する
		for result := range resultCh {
			fmt.Printf("Received: %d\n", result)
		}
		fmt.Println("Done receiving!")
	}
	resultCh := chanOwner() //　読み込み専用のコピーを受け取る
	consumer(resultCh)
	//for result := range resultCh {
	//	fmt.Printf("Received: %d\n", result)
	//}
	//fmt.Println("Done receiving!")
}
