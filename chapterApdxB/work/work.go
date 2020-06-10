package work

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "Type=Foo"

import (
	"fmt"
	"github.com/cheekybits/genny/generic"
)

type Type generic.Type // 算術不可なジェネリック型

func doWork(strings <-chan string) <-chan Type {
	completed := make(chan Type)
	go func() {
		defer fmt.Println("doWork exited.")
		defer close(completed)
		for s := range strings {
			fmt.Println(s)
		}
	}()
	return completed
}
