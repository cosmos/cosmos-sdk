package baseapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CircuitBreaker is an interface that defines the methods for a circuit breaker.
type CircuitBreakerInterface interface {
	IsAllowed(ctx sdk.Context, msg sdk.Msg) bool
}

// CircuitBreaker is a private implementation of the CircuitBreaker interface.
type CircuitBreaker struct {
	AllowedMsgs map[string]bool
}

func NewCircuitBreaker(msgs []sdk.Msg) *CircuitBreaker {
	cb := &CircuitBreaker{
		AllowedMsgs: make(map[string]bool),
	}

	for _, msg := range msgs {
		cb.AllowedMsgs[sdk.MsgTypeURL(msg)] = true
	}

	return cb
}

// IsAllowed returns whether the given message is allowed to be processed.
func (cb *CircuitBreaker) IsAllowed(ctx sdk.Context, msg sdk.Msg) bool {
	return cb.AllowedMsgs[sdk.MsgTypeURL(msg)]
}
