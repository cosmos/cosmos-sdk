# ADR 032: Change PubKey

## Changelog

- 30-09-2020: Initial Draft

## Status

Proposed

## Context

Currently, in the Cosmos SDK, the address of an auth account is always based on the hash of the public key.  Once an account is created, the public key for the account is set in stone, and cannot be changed.  This can be a problem for users, as key rotation is a useful security practice, but is not possible currently.  Furthermore, as multisigs are a type of pubkey, once a multisig for an account is set, it can not be updated.  This is problematic, as multisigs are often used by organizations or companies, who may need to change their set of multisig signers for internal reasons.

Transferring all the assets of an account to a new account with the updated pubkey is not sufficient, because some "engagements" of an account are not easily transferable.  For example, in staking, to transfer bonded Atoms, an account would have to unbond all delegations and wait the three week unbonding period.  Even more significantly, for validator operators, ownership over a validator is not transferrable at all, meaning that the operator key for a validator can never be updated, leading to poor operational security for validators. 

## Decision

We propose the creation of a new standaonle module called `changepubkey` that is an extension to `auth` that allows accounts to update the public key associated with their account, while keeping the address the same.

This is possible because the Cosmos SDK `StdAccount` stores the public key for an account in state, instead of making the assumption that the public key is included in the transaction (whether explicitly or implicitly through the signature) as in other blockchains such as Bitcoin and Ethereum.  Because the public key is stored on chain, it is okay for the public key to not hash to the address of an account, as the address is not pertinent to the signature checking process.

To build this system, we design a new Msg type as follows:

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

As an example, account pubkey change message can be defined as follows.

```json
{
    "type": "cosmos-sdk/StdTx",
    "value": {
        "address": "cosmos1wf5h7meplxu3sc6rk2agavkdsmlsen7rgsasxk",
        "public_key": "cosmospub1addwnpepqdszcr95mrqqs8lw099aa9h8h906zmet22pmwe9vquzcgvnm93eqygufdlv"
    },
    "signature": "a9n7pIqCUuYJTCm7ZBv1cqqlM3uYyX/7SnaSXA8zrG0CBWP6p55pTFFHYn39tVvFtRbGE7gXF1qCiaOilJ8NtQ=="
}
```

Here, the signature is signed for the public key thats current in-state for account `cosmos1wf5h7meplxu3sc6rk2agavkdsmlsen7rgsasxk`, as normally done in the ante-handler.

Once, approved, the handler for this message type, which takes in the AccountKeeper, will update the in-state pubkey for the account and replace it with the pubkey from the Msg.

Because an account can no longer be pruned from state once its pubkey has changed, we can charge an additional gas fee for this operation to compensate for this this externality (this bound gas amount is configured as parameter `PubKeyChangeCost`). The bonus gas is charged inside handler, using the `ConsumeGas` function.

```go
	amount := ak.GetParams(ctx).PubKeyChangeCost
	ctx.GasMeter().ConsumeGas(amount, "pubkey change fee")
```


## Consequences

### Positive

* Will allow users and validator operators to employ better operational security practices with key rotation.
* Will allow organizations or groups to easily change and add/remove multisig signers.

### Negative

Breaks the current assumed relationship between address and pubkeys as H(pubkey) = address. This has a couple of consequences.

* We cannot prune accounts with 0 balance that have had their pubkey changed (we currently do not currently do this anyways, but the reason we have account numbers is presumably for this purpose).
* This makes wallets that support this feature more complicated. For example, if an address on chain was updated, the corresponding key in the CLI wallet also needs to be updated.

### Neutral

* While the purpose of this is intended to allow the owner of an account to update to a new pubkey they own, this could technically also be used to transfer ownership of an account to a new owner.  For example, this could be use used to sell a staked position without unbonding or an account that has vesting tokens.  However, the friction of this is very high as this would essentially have to be done as a very specific OTC trade. Furthermore, additional constraints could be added to prevent accouns with Vesting tokens to use this feature.
* Will require that PubKeys for an account are included in the genesis exports.

## References

