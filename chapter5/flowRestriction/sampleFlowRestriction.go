package main

import (
	"context"
	"golang.org/x/time/rate"
	"log"
	"os"
	"sort"
	"sync"
	"time"
)

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

type RateLimiter interface {
	Wait(context.Context) error
	Limit() rate.Limit
}

type multiLimiter struct {
	limiters []RateLimiter
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit) // 最適化を実装して各RateLimiterのLimit()でソートする
	return &multiLimiter{limiters: limiters}
}

// 全ての子の流量資源のインスタンスを辿って、それぞれのWaitを呼び出している。
func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

// MultiLimiterでソートしているからここで返されるのは最も厳しい制限
// 最も長い待機時間を待つことが保証されている。一部しか待機しなかった場合、最も長く待機する制限に関して、再計算された結果の残り時間を待機するだけになる。
func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

type APIConnection struct {
	//rateLimiter RateLimiter
	networkLimit,
	diskLimit,
	apiLimit RateLimiter
}

func Open() *APIConnection {
	return &APIConnection{
		//rateLimiter: rate.NewLimiter(rate.Limit(1), 2),
		apiLimit: MultiLimiter(
			rate.NewLimiter(Per(2, time.Second), 2),
			rate.NewLimiter(Per(10, time.Minute), 10),
		),
		diskLimit: MultiLimiter(
			rate.NewLimiter(rate.Limit(1), 2),
		),
		networkLimit: MultiLimiter(
			rate.NewLimiter(Per(3, time.Second), 3),
		),
	}
}

// ファイルの読み込み
// このリクエストはネットワーク経由で行うので、context.Contextを最初の引数にとってリクエストをキャンセルしたり、サーバーに値を渡す必要がある場合に備える。
func (a *APIConnection) ReadFile(ctx context.Context) error {
	//if err := a.rateLimiter.Wait(ctx); err != nil {
	//	return err
	//}
	err := MultiLimiter(a.apiLimit, a.diskLimit).Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	//if err := a.rateLimiter.Wait(ctx); err != nil {
	//	return err
	//}
	err := MultiLimiter(a.apiLimit, a.networkLimit).Wait(ctx)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	defer log.Printf("Done.")
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	apiConnection := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile: %v", err)
			}
			log.Printf("ReadFile")
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ResolveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress: %v", err)
			}
			log.Printf("ResolveAddress")
		}()
	}

	wg.Wait()
}
