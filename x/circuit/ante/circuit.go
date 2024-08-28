package ante

import (
	"context"
	"errors"

	"cosmossdk.io/core/transaction"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CircuitBreaker is an interface that defines the methods for a circuit breaker.
type CircuitBreaker interface {
	IsAllowed(ctx context.Context, typeURL string) (bool, error)
}

// CircuitBreakerDecorator is an AnteDecorator that checks if the transaction type is allowed to enter the mempool or be executed
type CircuitBreakerDecorator struct {
	circuitKeeper CircuitBreaker
}

func NewCircuitBreakerDecorator(ck CircuitBreaker) CircuitBreakerDecorator {
	return CircuitBreakerDecorator{
		circuitKeeper: ck,
	}
}

// If you copy this as reference and your app has the authz module enabled, you must either:
// - recursively check for nested authz.Exec messages in this function.
// - or error early if a nested authz grant is found.
// The circuit AnteHandler handles this with baseapp's service router: https://github.com/cosmos/cosmos-sdk/issues/18632.
func (cbd CircuitBreakerDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if err := cbd.ValidateTx(ctx, tx); err != nil {
		return ctx, err
	}

	return next(ctx, tx, simulate)
}

func (cbd CircuitBreakerDecorator) ValidateTx(ctx context.Context, tx transaction.Tx) error {
	// loop through all the messages and check if the message type is allowed
	msgs, err := tx.GetMessages()
	if err != nil {
		return err
	}

	for _, msg := range msgs {
		isAllowed, err := cbd.circuitKeeper.IsAllowed(ctx, sdk.MsgTypeURL(msg))
		if err != nil {
			return err
		}

		if !isAllowed {
			return errors.New("tx type not allowed")
		}
	}
	return nil
}
