# ADR 028: Change PubKey

## Changelog

- 30-09-2020: Initial Draft

## Status

Proposed

## Context

This msg will update the public key associated with an account to a new public key, while keeping the same address.
		 
This can be used for purposes such as passing ownership of an account to a new key for security reasons or upgrading multisig signers.

## Decision

We create a module called `changepubkey` that handle all the actions related to pubkey change stuff.

We define the type as follows:

```protobuf
message MsgChangePubKey {
  bytes address = 1 [
    (gogoproto.casttype) = "github.com/cosmos/cosmos-sdk/types.AccAddress"
  ];
  bytes pub_key = 2 [
    (gogoproto.jsontag) = "public_key,omitempty", (gogoproto.moretags) = "yaml:\"public_key\""
  ];
}
```

As an example, msg that change pubkey of an accont can be defined as follows.
```json
{
    "type": "cosmos-sdk/StdTx",
    "value": {
        "address": "cosmos1wf5h7meplxu3sc6rk2agavkdsmlsen7rgsasxk",
        "public_key": "cosmospub1addwnpepqdszcr95mrqqs8lw099aa9h8h906zmet22pmwe9vquzcgvnm93eqygufdlv"
    }
}
```
Here `PubKey` is bech32-encoded one where it has prefix for secp256k1 or ed25519.

In addition, bonus gas amount for changing pubkey is configured as parameter `PubKeyChangeCost`.
```go
	amount := GetParams(ctx).PubKeyChangeCost
	ctx.GasMeter().ConsumeGas(amount, "pubkey change fee")
```
Bonus gas is paid inside handler, using `ConsumeGas` function.

## Consequences

### Positive

This can be used for purposes such as passing ownership of an account to a new key for security reasons or upgrading multisig signers.

### Negative

Breaks the current assumed relationship between address and pubkeys as H(pubkey) = address. This has a couple of consequences.
* We cannot prune accounts with 0 balance that have had their pubkey changed (we currently do not currently do this anyways, but the reason we have account numbers is presumably for this purpose).
* This makes wallets that support this feature more complicated. For example, if an address on chain was updated, the corresponding key in the CLI wallet also needs to be updated.

### Neutral

## References

