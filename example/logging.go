package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rixtrayker/demo-smpp/internal/response"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	logger := response.GetLogger(ctx)

	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < 400; i++ {
		wg.Add(1)
		go func(i int) {
			logger.Info(i)
			wg.Done()
		}(i)
		if i == 233 {
			cancel()
		}
	}

	wg.Wait()

	duration := time.Since(startTime)
	fmt.Printf("Duration: %d milliseconds\n", duration.Milliseconds())
}