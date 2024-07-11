package session

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)


func calculateBackoff(min, max time.Duration, factor float64, retries int) time.Duration {
	random := time.Duration(rand.Int63n(int64(max-min))) + min
	return time.Duration(math.Pow(factor, float64(retries))) * random
}

func (s *Session) connectRetry(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return errors.New("connectRetry stopped")
		default:
			if err := s.connectSessions(); err != nil {
				logrus.WithError(err).Error("Failed to reconnect")
				time.Sleep(2 * time.Second)
			} else {
				return nil
			}
		}
	}
}
