package handlers

import (
	"context"
	"errors"
)

type ZainHandler struct {
	Handler
}

func NewZainHandler(ctx context.Context, handlerFunc func()) (*ZainHandler, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	return &ZainHandler{
		Handler: Handler{
			ctx:         ctx,
			cancel:      cancel,
			handlerFunc: zainHandlerFunc,
		},
	}, nil
}

func zainHandlerFunc() {
	// implement the handlerFunc
}