package session

import (
	"math"
	"math/rand"
	"time"
)


func calculateBackoff(min, max time.Duration, factor float64, retries int) time.Duration {
	random := time.Duration(rand.Int63n(int64(max-min))) + min
	return time.Duration(math.Pow(factor, float64(retries))) * random
}