package handlers

import (
	"context"
	"errors"
	"sync"
)

type Handler struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     *sync.WaitGroup
	handlerFunc func()
}

type HandlerInterface interface {
	Handle()
}

func NewHandler(ctx context.Context, handlerFunc func()) (*Handler, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	return &Handler{
		ctx:         ctx,
		cancel:      cancel,
		handlerFunc: handlerFunc,
	}, nil
}

func (h *Handler) Handle() {
	defer h.wg.Done()
	select {
	case <-h.ctx.Done():
		return
	default:
		h.handlerFunc()
	}
}