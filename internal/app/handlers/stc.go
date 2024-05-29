package handlers

import (
	"context"
	"errors"
)

type STCHandler struct {
	Handler
}

func NewSTCHandler(ctx context.Context, handlerFunc func()) (*STCHandler, error) {
	if ctx == nil {
		return nil, errors.New("context cannot be nil")
	}
	ctx, cancel := context.WithCancel(ctx)
	return &STCHandler{
		Handler: Handler{
			ctx:         ctx,
			cancel:      cancel,
			handlerFunc: stcHandlerFunc,
		},
	}, nil
}


func stcHandlerFunc() {
	// implement the handlerFunc
}

