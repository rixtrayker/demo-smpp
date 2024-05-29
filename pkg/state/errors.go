package state

import "fmt"

// StateError represents an error related to the state machine.
type StateError struct {
	State   int
	Message string
}

// Error implements the error interface.
func (e *StateError) Error() string {
	return fmt.Sprintf("%s: %s", stateNames[e.State], e.Message)
}

// NewStateError creates a new StateError.
func NewStateError(state int, message string) *StateError {
	return &StateError{
		State:   state,
		Message: message,
	}
}