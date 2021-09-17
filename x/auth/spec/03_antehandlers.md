<!--
order: 3
-->

# AnteHandlers

The `x/auth` module presently has no transaction handlers of its own, but does expose the special `AnteHandler`, used for performing basic validity checks on a transaction, such that it could be thrown out of the mempool.
The `AnteHandler` can be seen as a set of decorators that check transactions within the current context, per [ADR 010](https://github.com/cosmos/cosmos-sdk/blob/v0.43.0-alpha1/docs/architecture/adr-010-modular-antehandler.md).

Note that the `AnteHandler` is called on both `CheckTx` and `DeliverTx`, as Tendermint proposers presently have the ability to include in their proposed block transactions which fail `CheckTx`.

## Decorators

The auth module provides `AnteDecorator`s that are recursively chained together into a single `AnteHandler` in the following order:

- `SetUpContextDecorator`: Sets the `GasMeter` in the `Context` and wraps the next `AnteHandler` with a defer clause to recover from any downstream `OutOfGas` panics in the `AnteHandler` chain to return an error with information on gas provided and gas used.

- `RejectExtensionOptionsDecorator`: Rejects all extension options which can optionally be included in protobuf transactions.

- `MempoolFeeDecorator`: Checks if the `tx` fee is above local mempool `minFee` parameter during `CheckTx`.

- `ValidateBasicDecorator`: Calls `tx.ValidateBasic` and returns any non-nil error.

- `TxTimeoutHeightDecorator`: Check for a `tx` height timeout.

- `ValidateMemoDecorator`: Validates `tx` memo with application parameters and returns any non-nil error.

- `ConsumeGasTxSizeDecorator`: Consumes gas proportional to the `tx` size based on application parameters.

- `DeductFeeDecorator`: Deducts the `FeeAmount` from first signer of the `tx`. If the `x/feegrant` module is enabled and a fee granter is set, it will deduct fees from the fee granter account.

- `SetPubKeyDecorator`: Sets the pubkey from a `tx`'s signers that does not already have its corresponding pubkey saved in the state machine and in the current context.

- `ValidateSigCountDecorator`: Validates the number of signatures in `tx` based on app-parameters.

- `SigGasConsumeDecorator`: Consumes parameter-defined amount of gas for each signature. This requires pubkeys to be set in context for all signers as part of `SetPubKeyDecorator`.

- `SigVerificationDecorator`: Verifies all signatures are valid. This requires pubkeys to be set in context for all signers as part of `SetPubKeyDecorator`.

- `IncrementSequenceDecorator`: Increments the account sequence for each signer to prevent replay attacks.
