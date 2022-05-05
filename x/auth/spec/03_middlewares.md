<!--
order: 3
-->

# Middlewares

The `x/auth` module presently has no transaction handlers of its own, but does expose middlewares directly called from BaseApp's `CheckTx` and `DeliverTx`, which can be used for performing any operations on transactions, such as basic validity checks on a transaction such that it could be thrown out of the mempool, or routing the transactions to their `Msg` service to perform state transitions.
The middlewares can be seen as a set of decorators wrapped one on top of the other, that check transactions within the current context, per [ADR-045](https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-beta2/docs/architecture/adr-045-check-delivertx-middlewares.md).

Note that the middlewares are called on both `CheckTx` and `DeliverTx`, as Tendermint proposers presently have the ability to include in their proposed block transactions which fail `CheckTx`.

## List of Middleware

The auth module provides:

- one `tx.Handler`, called `RunMsgsTxHandler`, which routes each `sdk.Msg` from a transaction to the correct module `Msg` service, and runs each `sdk.Msg` to perform state transitions,
- a set of middlewares that are recursively chained together around the base `tx.Handler` in the following order (the first middleware's `pre`-hook is run first, and `post`-hook is run last):

  - `NewTxDecoderMiddleware`: Decodes the transaction bytes from ABCI `CheckTx` and `DeliverTx` into the SDK transaction type. This middleware is generally called first, as most middlewares logic rely on a decoded SDK transaction.
  - `GasTxMiddleware`: Sets the `GasMeter` in the `Context`.
  - `RecoveryTxMiddleware`: Wraps the next middleware with a defer clause to recover from any downstream panics in the middleware chain to return an error with information on gas provided and gas used.
  - `RejectExtensionOptionsMiddleware`: Rejects all extension options which can optionally be included in protobuf transactions.
  - `IndexEventsTxMiddleware`: Choose which events to index in Tendermint. Make sure no events are emitted outside of this middleware.
  - `ValidateBasicMiddleware`: Calls `tx.ValidateBasic` and returns any non-nil error.
  - `TxTimeoutHeightMiddleware`: Check for a `tx` height timeout.
  - `ValidateMemoMiddleware`: Validates `tx` memo with application parameters and returns any non-nil error.
  - `ConsumeGasTxSizeMiddleware`: Consumes gas proportional to the `tx` size based on application parameters.
  - `DeductFeeMiddleware`: Deducts the `FeeAmount` from first signer of the `tx`. If the `x/feegrant` module is enabled and a fee granter is set, it deducts fees from the fee granter account.
  - `SetPubKeyMiddleware`: Sets the pubkey from a `tx`'s signers that does not already have its corresponding pubkey saved in the state machine and in the current context.
  - `ValidateSigCountMiddleware`: Validates the number of signatures in the `tx` based on app-parameters.
  - `SigGasConsumeMiddleware`: Consumes parameter-defined amount of gas for each signature. This requires pubkeys to be set in context for all signers as part of `SetPubKeyMiddleware`.
  - `SigVerificationMiddleware`: Verifies all signatures are valid. This requires pubkeys to be set in context for all signers as part of `SetPubKeyMiddleware`.
  - `IncrementSequenceMiddleware`: Increments the account sequence for each signer to prevent replay attacks.
  - `WithBranchedStore`: Creates a new MultiStore branch, discards downstream writes if the downstream returns error.
  - `ConsumeBlockGasMiddleware`: Consume block gas.
  - `TipMiddleware`: Transfer tips to the fee payer in transactions with tips.

This default list of middlewares can be instantiated using the `NewDefaultTxHandler` function. If a chain wants to tweak the list of middlewares, they can create their own `NewTxHandler` function using the same template as `NewDefaultTxHandler`, and chain new middlewares in the `ComposeMiddleware` function.
