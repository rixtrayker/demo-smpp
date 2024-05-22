package config

import "golang.org/x/time/rate"

func NewRateLimiter(limit int) *rate.Limiter {
    return rate.NewLimiter(rate.Limit(limit), limit)
}
