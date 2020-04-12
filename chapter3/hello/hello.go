package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup

	sayHello := func() {
		defer wg.Done()
		fmt.Println("hello")
	}
	wg.Add(1)
	go sayHello()
	wg.Wait()

	for _, salutation := range []string{"hello", "greetings", "good days"} {
		wg.Add(1)
		go func(salutation string) {
			defer wg.Done()
			fmt.Println(salutation)
			//}()
		}(salutation)
	}
	wg.Wait()
}
