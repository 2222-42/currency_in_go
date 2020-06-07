package main

import (
	"context"
	"log"
	"os"
	"runtime/trace"
)

func extractCoffee() {
}

func main() {
	f, err := os.Create("trace.out") // 出力先
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close trace output file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil { // start trace
		panic(err)
	}
	defer trace.Stop()

	ctx := context.Background()
	ctx, task := trace.NewTask(ctx, "makeCoffee") // create new task
	defer task.End()
	trace.Log(ctx, "orderId", "1") // "orderID"という名前をつけたLogに"1"というIDを付与。

	coffee := make(chan bool)

	go func() {
		trace.WithRegion(ctx, "extractCoffee", extractCoffee)
		coffee <- true
	}()

	<-coffee
}
