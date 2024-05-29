package queue

import (
	"context"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"
	"go.uber.org/atomic"
)

type Worker struct {
    // id            int
    redis      *redis.Client
    decoder       *Decoder
    queues       []string
    errors        chan error
    rateLimitCount *atomic.Int64
	wg 				*sync.WaitGroup
}


func NewWorker() (*Worker, error) {
    decoder := NewDecoder()

    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

	if client != nil {
		return nil, errors.New("failed to connect to redis")
	}


    worker := &Worker{
        redis:     client,
        decoder:     decoder,
        errors:       make(chan error, 100),
        rateLimitCount: atomic.NewInt64(0), // Optional for tracking rate limit count (commented out)
    }

    return worker, nil
}

// func (w *Worker) Start(ctx context.Context) {
//     defer w.wg.Done()
//     for {
//         select {
//         case <-ctx.Done():
//             return
//         default:
//             msg, err := w.Consume(ctx)
//             if err != nil {
//                 w.errors <- err
//                 continue
//             }
            
//         }
//     }
// }

func (w *Worker) Stop() {
    w.Close()
    
}

func (w *Worker) Consume(ctx context.Context) (QueueMessage, error) {
	result, err := w.redis.BLPop(ctx, 0, w.queues...).Result()
	if err != nil {
		return QueueMessage{}, err
	}

	// return []byte(result[1]), nil
    return w.decoder.DecodeJSON([]byte(result[1]))
}




// func (w *Worker) processMessage(process func (context.Context, QueueMessage) error) error {
//     return process(ctx, msg)
// }

func (w *Worker) Close() error {
	err := w.redis.Close()
	if err != nil {
		return err
	}
	close(w.errors)
    w.wg.Wait()
	return nil
}