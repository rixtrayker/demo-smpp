package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/rixtrayker/demo-smpp/internal/response"
)

func main() {
	logger := response.GetLogger()

	var wg sync.WaitGroup

	startTime := time.Now()

	for i := 0; i < 400; i++ {
		wg.Add(1)
		go func(i int) {
			logger.Info(i)
			wg.Done()
		}(i)
	}

	wg.Wait()

	duration := time.Since(startTime)
	fmt.Printf("Duration: %d milliseconds\n", duration.Milliseconds())
}