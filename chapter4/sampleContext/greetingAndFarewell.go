package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup

	//done := make(chan interface{})
	//defer close(done)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel() // errがある場合、mainがContextをキャンセルするようにする。
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
			//return
		}
	}()

	wg.Wait()
}

func printGreeting(ctx context.Context) error {
	//func printGreeting(done <-chan interface{}) error {
	greeting, err := genGreeting(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func genGreeting(ctx context.Context) (string, error) {
	//func genGreeting(done <-chan interface{}) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second) // 独自のcontext.Contextを作成する。一秒後に戻されたコンテキストを自動的にキャンセル、当然このContextを使うlocaleもキャンセルされる
	defer cancel()

	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func printFarewell(ctx context.Context) error {
	//func printFarewell(done <-chan interface{}) error {
	greeting, err := genFarewell(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func genFarewell(ctx context.Context) (string, error) {
	//func genFarewell(done <-chan interface{}) (string, error) {
	//	ctx, cancel := context.WithTimeout(ctx, 1*time.Second) // 一秒後に戻されたコンテキストを自動的にキャンセル、当然このContextを使うlocaleもキャンセルされる
	//	defer cancel()

	switch locale, err := locale(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(ctx context.Context) (string, error) {
	//func locale(done <-chan interface{}) (string, error) {

	if deadline, ok := ctx.Deadline(); ok { // deadlineが設定されているか、設定されていてその時刻を過ぎているのなら、DeadlineExceededエラーを返す
		if deadline.Sub(time.Now().Add(1*time.Minute)) <= 0 {
			return "", context.DeadlineExceeded
		}
	}
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(1 * time.Minute):
	}
	return "EN/US", nil
}
