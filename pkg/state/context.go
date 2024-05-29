package state

import "context"

// CreateClientContext creates a new context with cancellation for a client.
func (s *State) CreateClientContext() (context.Context, context.CancelFunc) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	ctx, cancel := context.WithCancel(s.ctx)
	s.cancels["client"] = &cancel
	return ctx, cancel
}

// CancelClientContext cancels the context for a client.
func (s *State) CancelClientContext() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if cancel, ok := s.cancels["client"]; ok {
		(*cancel)()
		delete(s.cancels, "client")
	}
}