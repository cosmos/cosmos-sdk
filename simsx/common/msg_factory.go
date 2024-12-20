package common

import (
	"context"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SimMsgFactoryX is an interface for creating and handling fuzz test-like simulation messages in the system.
type SimMsgFactoryX interface {
	// MsgType returns an empty instance of the concrete message type that the factory provides.
	// This instance is primarily used for deduplication and reporting purposes.
	// The result must not be nil
	MsgType() sdk.Msg

	// Create returns a FactoryMethod implementation which is responsible for constructing new instances of the message
	// on each invocation.
	Create() FactoryMethod

	// DeliveryResultHandler returns a SimDeliveryResultHandler instance which processes the delivery
	// response error object. While most simulation factories anticipate successful message delivery,
	// certain factories employ this handler to validate execution errors, thereby covering negative
	// test scenarios.
	DeliveryResultHandler() SimDeliveryResultHandler
}

type (
	// FactoryMethod is a method signature implemented by concrete message factories for SimMsgFactoryX
	//
	// This factory method is responsible for creating a new `sdk.Msg` instance and determining the
	// proposed signers who are expected to successfully sign the message for delivery.
	//
	// Parameters:
	// - ctx: The context for the operation
	// - testData: A pointer to a `ChainDataSource` which provides helper methods and simple access to accounts
	//   and balances within the chain.
	// - reporter: An instance of `SimulationReporter` used to report the results of the simulation.
	//   If no valid message can be provided, the factory method should call `reporter.Skip("some comment")`
	//   with both `signer` and `msg` set to nil.
	//
	// Returns:
	// - signer: A slice of `SimAccount` representing the proposed signers.
	// - msg: An instance of `sdk.Msg` representing the message to be delivered.
	FactoryMethod func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg sdk.Msg)

	// FactoryMethodWithFutureOps extended message factory method for the gov module or others that have to schedule operations for a future block.
	FactoryMethodWithFutureOps[T sdk.Msg] func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter, fOpsReg FutureOpsRegistry) ([]SimAccount, T)

	// FactoryMethodWithDeliveryResultHandler extended factory method that can return a result handler, that is executed on the delivery tx error result.
	// This is used in staking for example to validate negative execution results.
	FactoryMethodWithDeliveryResultHandler[T sdk.Msg] func(ctx context.Context, testData *ChainDataSource, reporter SimulationReporter) (signer []SimAccount, msg T, handler SimDeliveryResultHandler)
)
