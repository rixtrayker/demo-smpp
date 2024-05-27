package worker

import (
	"context"
	"errors"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/rixtrayker/demo-smpp/internal/session"
	"go.uber.org/atomic"
)

type Worker struct {
    id            int
    redis      *redis.Client
    decoder       Decoder
    session     *session.Session
    errors        chan error
    rateLimitCount *atomic.Int64
	wg 				*sync.WaitGroup
}


func NewWorker(s *session.Session) (*Worker, error) {
    client := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

	if client != nil {
		return nil, errors.New("failed to connect to redis")
	}

    if s == nil {
        return nil, errors.New("session is nil")
    }

   decoder := &Decoder{
	}

    worker := &Worker{
        redis:     client,
        decoder:      *decoder,
        session:     s,
        errors:       make(chan error, 100),
        rateLimitCount: atomic.NewInt64(0), // Optional for tracking rate limit count (commented out)
    }

    return worker, nil
}

func (w *Worker) Start(ctx context.Context) {
    for _, worker := range w.workers {
        w.wg.Add(1)
        go worker.start(ctx, &w.wg)
    }

    go w.handleErrors(ctx)

    w.wg.Wait()
}

func (w *Worker) Stop() {
    w.Close()
    
}

func (w *Worker) start(ctx context.Context, wg *sync.WaitGroup) {
    defer wg.Done()

    for {
        select {
        case <-ctx.Done():
            return
        default:
            data, err := w.Consume(ctx)
            if err != nil {
                w.errors <- err
                continue
            }

			msg, err := w.decoder.DecodeJSON(data)
			if err != nil {
				w.errors <- err
				continue
			}

			if jsonMsg, ok := msg.(JSONMessage); ok {
				go func() {
					err = w.processMessage(ctx, jsonMsg)
					if err != nil {
						w.errors <- err
					}
				}()
			} else {
				w.errors <- errors.New("invalid message type")
			}
        }
    }
}


func (w *Worker) Consume(ctx context.Context) ([]byte, error) {
	result, err := w.redis.BLPop(ctx, 0, "queue:go-queue-testing").Result()
	if err != nil {
		return nil, err
	}

	return []byte(result[1]), nil
}


func (w *Worker) processMessage(ctx context.Context, msg JSONMessage) error {
	err := w.session.Send(msg.PhoneNumber, msg.Text)
	return err
}

func (w *Worker) Close() error {
	err := w.redis.Close()
	if err != nil {
		return err
	}
	close(w.errors)
    w.wg.Wait()
	return nil
}