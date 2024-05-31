package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rixtrayker/demo-smpp/internal/dtos"
	"github.com/rixtrayker/demo-smpp/internal/response"
)


func main(){
	ctx := context.Background()
	rw := response.NewResponseWriter(ctx)
	data := &dtos.ReceiveLog{
		MobileNo: "43422442",
		MessageID: "42334232",
	}
	now := time.Now()

	wg := sync.WaitGroup{}
	wg.Add(1000000)
	for j := 0; j < 10; j++ {
		go func() {
			for i := 0; i < 100000; i++ {
				go func(data *dtos.ReceiveLog) {
					defer wg.Done()
				// defer wg.Done()
					rw.WriteResponse(data)
				}(data)
			}
		}()
	}

	wg.Wait()
	fmt.Println(time.Since(now).Seconds())
	// total time in seconds to process 1,000,000 records
}