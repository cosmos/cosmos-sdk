<!--
order: 15
-->

# Transaction Tips

Transaction tips are a mechanism to pay for transaction fees using another denom than the native fee denom of the chain. {synopsis}

## Context

In a Cosmos ecosystem where more and more chains are connected via [IBC](https://ibc.cosmos.network/), it happens that users want to perform actions on chains where they don't have native tokens yet. An example would be an Osmosis user who wants to vote on a proposal on the Cosmos Hub, but they don't have ATOMs in their wallet. A solution would be to swap OSMO for ATOM just for voting on this proposal, but that is cumbersome. Cross-chain DeFi project [Emeris](https://emeris.com/) is another use case.

Transaction tips is a new solution for cross-chain transaction fees payment, whereby the transaction initiator signs a transaction without specifying fees, but uses a new `Tip` field. They send this signed transaction to a fee relayer who will broadcast the final transaction, and the SDK provides a mechanism that will transfer the pre-defined `Tip` to the fee payer, to cover for fees.

Assuming we have two chains, A and B, we define the following terms:

- **the tipper**: this is the initiator of the transaction, who wants to execute a `Msg` on chain A, but doesn't have any native chain A tokens, only chain B tokens. In our example above, the tipper is the Osmosis (chain B) user wanting to vote on a Cosmos Hub (chain A) proposal.
- **the fee payer**: this is the party that will relay and broadcast the final transaction on chain A, and has chain A tokens. The tipper doesn't need to trust the feepayer.
- **the target chain**: the chain where the `Msg` is executed, chain A in this case.

## Transaction Tips Flow

The transaction tips flow happens in multipe steps.

1. The tipper sends via IBC some chain B tokens to chain A. These tokens will be used to pay for fees on the target chain A. This means that chain A's bank module holds some IBC tokens under the tipper's address.

2. The tipper drafts a transaction to be executed on the chain A. It can include chain A `Msg`s. However, instead of creating a normal transaction, they create the following `SignDocDirectAux` document:

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-beta1/proto/cosmos/tx/v1beta1/tx.proto#L67-L93

where `Tip` is defined as

+++ https://github.com/cosmos/cosmos-sdk/blob/v0.46.0-beta1/proto/cosmos/tx/v1beta1/tx.proto#L219-L228

Notice that this document doesn't sign over the final chain A fees. Instead, it includes a `Tip` field. It also doesn't include the whole `AuthInfo` object as in `SIGN_MODE_DIRECT`, only the minimum information needed by the tipper

3. The tipper signs the `SignDocDirectAux` document and attaches the signature to it, then sends the signed document to the fee payer.

4. From the signed `SignDocDirectAux` document, the fee payer constructs a transaction, using the following algorithm:

- use as `TxBody` the exact `SignDocDirectAux.body_bytes`, to not alter the original intent of the tipper,
- create an `AuthInfo` with:
  - `AuthInfo.Tip` copied from `SignDocDirectAux.Tip`,
  - `AuthInfo.Fee` chosen by the fee payer, which should cover for the transaction gas, but also be small enough so that the tip/fee exchange rate is economically interesting for the fee payer,
  - `AuthInfo.SignerInfos` has two signers: the first signer is the tipper, using the public key, sequence and sign mode specified in `SignDocDirectAux`; and the second signer is the fee payer, using their favorite sign mode,
- a `Signatures` array with two items: the tipper's signature from `SignDocDirectAux`, and the final fee payer's signature.

5. Broadcast the final transaction signed by the two parties to the target chain. Once included, the Cosmos SDK will trigger a transfer of the `Tip` specified in the transaction from the tipper address to the fee payer address.

### Fee Payers Market

The benefit of transaction tips for the tipper is clear: there is no need to swap tokens before executing a cross-chain message.

For the fee payer, the benefit is in the tip v.s. fee exchange. Put simply, the fee payer pays the fees of an unknown tipper's transaction, and gets in exchange the tip that the tipper chose. There is an economic incentive for the fee payer to do so only when the tip is greater than the transaction fees, given the exchange rates between the two tokens.

In the future, we imagine a market where fee payers will compete to include transactions from tippers, who on their side will optimize by specifying the lowest tip possible. A number of automated services might spin up to perform transaction gas simulation and exchange rate monitoring to optimize both the tip and fee values in real-time.

### Tipper and Fee Payer Sign Modes

As we mentioned in the flow above, the tipper signs over the `SignDocDirectAux`, and the fee payer signs over the whole final transaction. As such, both parties might use different sign modes.

- The tipper MUST use `SIGN_MODE_DIRECT_AUX` or `SIGN_MODE_LEGACY_AMINO_JSON`. That is because the tipper needs to sign over the body, the tip, but not the other signers' information.
- The fee payer MUST use `SIGN_MODE_DIRECT` or `SIGN_MODE_LEGACY_AMINO_JSON`. The fee payer signs over the whole transaction.

For example, if the fee payers signs the whole transaction with `SIGN_MODE_DIRECT_AUX`, it will be rejected by the node, as that would introduce malleability issues (`SIGN_MODE_DIRECT_AUX` doesn't sign over fees).

In both cases, using `SIGN_MODE_LEGACY_AMINO_JSON` is recommended only if hardware wallet signing is needed.

## Enabling Tips on your Chain

The transaction tips functionality is introduced in Cosmos SDK v0.46, so earlier versions do not have support for tips. If you're using v0.46 or later, then enabling tips on your chain is as simple as adding the `TipMiddleware` in your middleware stack:

```go
// NewTxHandler defines a TxHandler middleware stack.
func NewTxHandler(options TxHandlerOptions) (tx.Handler, error) {
    // --snip--

    return ComposeMiddlewares(
        // base tx handler that executes Msgs
        NewRunMsgsTxHandler(options.MsgServiceRouter, options.LegacyRouter),
        // --snip other middlewares--

        // Add the TipMiddleware
        NewTipMiddleware(options.BankKeeper),
    )
}
```

Notice that `NewTipMiddleware` needs a reference to the BankKeeper, for transferring the tip to the fee payer.

If you are using the Cosmos SDK's default middleware stack `NewDefaultTxHandler()`, then the tip middleware is included by default.

## CLI Usage

The Cosmos SDK also provides some CLI tooling for the transaction tips flow, both for the tipper and for the feepayer.

For the tipper, the CLI has two new flags: `--aux` and `--tip`. The `--aux` flag is used to denote that we are creating a `SignDocDirectAux`.
