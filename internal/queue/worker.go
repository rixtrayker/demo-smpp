package queue

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"
	"go.uber.org/atomic"
)

type Worker struct {
    // id            int
    ctx          context.Context
    redis      *redis.Client
    decoder       *Decoder
    queues       []string
    errors        chan error
    rateLimitCount *atomic.Int64
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
        ctx: context.Background(),
        redis:     client,
        decoder:     decoder,
        errors:       make(chan error, 100),
        rateLimitCount: atomic.NewInt64(0), // Optional for tracking rate limit count (commented out)
    }

    return worker, nil
}

// func (w *Worker) Start(ctx context.Context) {
//     for {
//         select {
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

func (w *Worker) Consume(q []string) (QueueMessage, error) {

    result, err := w.redis.BLPop(w.ctx, 0, q...).Result()
	if err != nil {
		return QueueMessage{}, err
	}

	// return []byte(result[1]), nil
    return w.decoder.DecodeJSON([]byte(result[1]))
}

func (w *Worker) streamQueueMessage(q []string) (<-chan QueueMessage, <-chan error) {
    messages := make(chan QueueMessage, 200) // Buffered channel with size 200
    errors := make(chan error)

    go func() {
        defer close(messages)
        defer close(errors)

        for {
            select {
            case <-w.ctx.Done():
                return
            default:
                result, err := w.Consume(q)
                if err != nil {
                    errors <- err
                    continue
                }

                messages <- result
            }
        }
    }()
    
    return messages, errors
}

func (w *Worker) Stream(q []string) (<-chan MessageData, <-chan error) {
    var queues []string

    if len(q) == 0 {
        queues = w.queues
    } else {
        queues = q
    }

    messages, errors := w.streamQueueMessage(queues)
    data := make(chan MessageData)
    go func() {
        defer close(data)
        for msgQ := range messages {
            for _, msg := range msgQ.Deflate() {
                    data <- msg
            }
        }
    }()

    return data, errors
}

func (w *Worker) Push(ctx context.Context, queue string, message *MessageData) error {
    _, err := w.redis.RPush(ctx, queue, message).Result()
    if err != nil {
        return err
    }
    return nil
}

func (w *Worker) Close() error {
	err := w.redis.Close()
	if err != nil {
		return err
	}
	close(w.errors)
	return nil
}