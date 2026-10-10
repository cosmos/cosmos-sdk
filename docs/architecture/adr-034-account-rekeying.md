# ADR 034: Account Rekeying

## Changelog

* 30-09-2020: Initial Draft
* 09-10-2026: Revised against the current codebase. Adds proof of possession for the new key, the ante handler rule, rotation history and reverse-index state, queries, client support, and the places that derive an address from a pubkey. Motivated by multisig membership changes and migration to ML-DSA-65.

## Status

PROPOSED

## Abstract

Account rekeying is a process that allows an account to replace its authentication pubkey with a new one while keeping its address.

## Context

Currently, in the Cosmos SDK, the address of an auth `BaseAccount` is based on the hash of the public key.  Once an account is created, the public key for the account is set in stone, and cannot be changed.  This can be a problem for users, as key rotation is a useful security practice, but is not possible currently.  Furthermore, as multisigs are a type of pubkey, once a multisig for an account is set, it cannot be updated.  This is problematic, as multisigs are often used by organizations or companies, who may need to change their set of multisig signers for internal reasons.

Transferring all the assets of an account to a new account with the updated pubkey is not sufficient, because some "engagements" of an account are not easily transferable.  For example, in staking, to transfer bonded Atoms, an account would have to unbond all delegations and wait the three-week unbonding period.  Even more significantly, for validator operators, ownership over a validator is not transferable at all, meaning that the operator key for a validator can never be updated, leading to poor operational security for validators.

### Context (2026)

The original draft assumed that, because the pubkey is stored in state, nothing in the SDK cares whether it hashes to the address. That is no longer true, and several things have changed since 2020:

* The ante handler enforces `hash(pubkey) == address`. `SetPubKeyDecorator` (`x/auth/ante/sigverify.go`, line 100) rejects any tx whose signer pubkey does not hash to the signer address, even when the account already has a pubkey stored. Signature verification itself (`SigVerificationDecorator`) already uses the stored pubkey, so only the `SetPubKeyDecorator` check stands in the way.
* `BaseAccount.Validate` (`x/auth/types/account.go`), the default mempool's signer extraction (`types/mempool/signer_extraction_adapter.go`, `types/mempool/sender_nonce.go`) and client signing (`client/tx/tx.go`) all derive an account address from a pubkey.
* ML-DSA-65 (FIPS 204) is now supported for account keys ([#26472](https://github.com/cosmos/cosmos-sdk/pull/26472)). Existing secp256k1 accounts, and validator operators in particular, need a way to move to a post-quantum key without moving funds or unbonding. Multisigs that mix secp256k1 and ML-DSA-65 keys are a natural migration step, and they too need membership changes.
* Validator consensus key rotation has shipped in `x/staking` (`MsgRotateConsPubKey`). It covers consensus keys only. The operator account key remains fixed, which is the gap this ADR closes.

## Alternatives

* **Move funds to a new account.** Doesn't work for bonded stake, vesting schedules, validator ownership or anything else keyed by address (see Context).
* **Rekey without a proof of possession.** This is the original draft's design. A typo or wrong key permanently locks the account, including staked funds and validator ownership, and an account can bind itself to a key someone else controls, which confuses attribution. This revision rejects that design.
* **Allow rekeying through authz, group or gov.** This would let a grantee or a module take over the granter's account. This revision rejects it as well (decision 4).

## Decision

We add a feature to `x/auth` that lets an account replace the public key associated with it while its address stays the same.

This works because `BaseAccount` stores the account's public key in state. Signature verification uses the stored key, so the key does not need to hash to the address once it is set.

### 1. Message

```protobuf
service Msg {
  rpc ChangePubKey(MsgChangePubKey) returns (MsgChangePubKeyResponse);
}

message MsgChangePubKey {
  option (cosmos.msg.v1.signer) = "address";
  option (amino.name)           = "cosmos-sdk/MsgChangePubKey";

  string              address     = 1 [(cosmos_proto.scalar) = "cosmos.AddressString"];
  google.protobuf.Any new_pub_key = 2 [(cosmos_proto.accepts_interface) = "cosmos.crypto.PubKey"];
  // proof is the proto encoding of cosmos.tx.signing.v1beta1.SignatureDescriptor.Data.
  bytes proof = 3;
}

message MsgChangePubKeyResponse {}
```

`MsgChangePubKey` is signed by the account's current key, like any other tx from the account. `proof` holds the proto encoding of `cosmos.tx.signing.v1beta1.SignatureDescriptor.Data`. It is kept as opaque bytes to avoid putting a oneof through the aminojson encoder.

### 2. Proof of possession is required

A wrong key means permanent loss of the account, including staked funds and validator ownership. Binding an account to a key you don't control also confuses attribution. The new key must therefore sign a proof.

The proof is signed over an [ADR-036](./adr-036-arbitrary-signature.md) amino-JSON document, so hardware wallets such as Ledger can produce it:

* `signer` is the bech32 encoding of `new_pub_key.Address()`.
* `data` is the deterministic proto encoding of:

```protobuf
message ChangePubKeyProofDoc {
  string              chain_id       = 1;
  uint64              account_number = 2;
  string              address        = 3;
  google.protobuf.Any new_pub_key    = 4;
}
```

The resulting sign bytes are the sorted JSON:

```json
{"account_number":"0","chain_id":"","fee":{"amount":[],"gas":"0"},"memo":"","msgs":[{"type":"sign/MsgSignData","value":{"data":"<base64 proto(doc)>","signer":"<bech32 new_pub_key.Address()>"}}],"sequence":"0"}
```

ADR-036 leaves replay protection to the application. Here, `chain_id`, `account_number` and `address` inside the doc bind the proof to one account on one chain, so a proof made for one account or chain is rejected for another.

The doc has no freshness component such as the account sequence or the current pubkey, so proofs are reusable by design. A proof made for key K on account A stays valid after A rotates away from K, and the account's current key holder can later rotate A back to K without asking K's owner to sign again. The impact is low: only the current key holder can submit `MsgChangePubKey`, and the proof only shows that K's owner once agreed to control A. Wallets that care about this should treat a past proof as standing consent.

For a single key, the proof is a `SingleSignatureData` verified with `new_pub_key.VerifySignature`. For a multisig new key, the proof is a `MultiSignatureData` that meets the new key's threshold, verified with `LegacyAminoPubKey.VerifyMultisignature`. Every signature in the proof, including each one inside a multisignature, must use `SIGN_MODE_LEGACY_AMINO_JSON`, since the sign bytes are an amino-JSON document; any other sign mode is rejected. A multisignature proof may nest at most 2 `MultiSignatureData` levels, matching the limit on the new key's multisig nesting.

### 3. Accepted new key types

The accepted types match what `DefaultSigVerificationGasConsumer` accepts:

* `secp256k1`, `secp256r1`, `ed25519`, `mldsa65`;
* `LegacyAminoPubKey` (multisig) built from those types, except that `secp256r1` is accepted only as a top-level key. The amino codec `LegacyAminoPubKey.Address()` uses does not register `secp256r1`, so a multisig holding one cannot compute its address.

A multisig may contain multisigs, but those may not: at most 2 levels of multisig (`MaxSignatureTreeDepth`), and at most 32 direct subkeys per multisig (`MaxSignatureTreeBreadth`). These are the limits the ante handler's `flattenSignatures` enforces on every tx signature, and both it and the rekey check use the same constants. Without them an account could rotate to, e.g., a 1-of-1 of a 1-of-1 of a 1-of-1, whose signatures the ante handler always rejects, and then could never sign again, not even to rotate back.

`CountSubKeys(new_pub_key) <= Params.TxSigLimit` must hold, so the account can't lock itself out by adopting a key it can never sign with.

The handler rejects `ModuleCredential`, `secp256k1eth`, a nil key, and a key that `Equals` the current one.

### 4. Who can rekey

An account can rekey only if all of the following hold:

* it is not an `sdk.ModuleAccountI`;
* its stored pubkey is not nil when the handler runs;
* its stored pubkey is not a `*ModuleCredential`.

That excludes gov, group policies and other module-derived accounts. The non-nil rule also excludes accounts that never had a key, such as CosmWasm contract accounts and ibc-go interchain accounts. Those are not `ModuleAccountI` and have a nil stored pubkey, but they can still dispatch messages as themselves (a contract forwarding `Any` or Stargate messages, or an ICA host whose `AllowMessages` includes `MsgChangePubKey`). Without the rule, such a dispatch could bind the account to an attacker's key, and the attacker could then sign txs for it directly, bypassing the contract logic or controller. Legitimate user signers lose nothing: `SetPubKeyDecorator` stores the signer's pubkey earlier in the same tx, so by the time the handler runs their stored pubkey is set.

`x/authz` refuses to grant `MsgChangePubKey` and refuses to dispatch it, including under a `GenericAuthorization` created before this feature existed. A grant would let the grantee take over the granter's account.

### 5. Ante handler rule

In `SetPubKeyDecorator`:

* If the signer account already has a stored pubkey, a non-nil tx pubkey must `Equals` it. Otherwise the tx fails with `ErrInvalidPubKey`. A nil tx pubkey is allowed, as today.
* If no pubkey is stored, the existing `pk.Address() == signer` check still applies, and the pubkey is set on the account as today.

Both checks run under the same conditions as today's address check: only when `simulate` is false and `ctx.IsSigverifyTx()` is true. In simulate mode a nil tx pubkey is replaced with a placeholder secp256k1 key that would not `Equals` a rekeyed account's stored key, so applying the check there would break gas estimation. The checks do run in `CheckTx`, `RecheckTx` and `FinalizeBlock`, because `SetPubKeyDecorator` is not skipped on recheck.

`SigVerificationDecorator` already verifies signatures against the stored pubkey, so it needs no change.

In simulate mode, `ConsumeTxSizeGasDecorator` sizes each missing signature from the signer's stored pubkey (falling back to the tx pubkey, then to the secp256k1 placeholder), so a gas estimate for an account rekeyed to an ML-DSA-65 key, or to a multisig of them, covers the larger signature. Gas outside simulate mode is unchanged.

Txs signed by the old key fail once the rotation is committed. This covers an ordered tx with sequence `n+1` that was signed before the rotation, and unordered txs. Such txs fail in `FinalizeBlock` with no state change. An old-key tx that carries its pubkey fails the `Equals` check on the first `RecheckTx` after the rotation and is evicted from the mempool. An old-key tx that omits its pubkey passes `SetPubKeyDecorator`, and `SigVerificationDecorator` skips signature verification on recheck, so it may sit in the mempool until it is included and fails or until it expires. This is acceptable because it cannot succeed. Operators and wallets should expect such txs.

### 6. State

Two new collections in the `acc` store:

* `PubKeyHistory collections.Map[collections.Pair[sdk.AccAddress, uint64], types.PubKeyHistoryEntry]`, at prefix `91`. It is keyed by (account, rotation index). The rotation index is a per-account counter: the account's first rotation is `0`, and each later rotation uses one more than the highest index stored for that account. The value holds the replaced pubkey, the block height and block time at which it was replaced, and the new key's natural address. This lets clients find which key was active at a given time, for example to verify timestamped off-chain signatures.
* `RekeyIndex collections.KeySet[collections.Pair[sdk.AccAddress, sdk.AccAddress]]`, at prefix `92`. It holds (natural address of the current key, account), and only while the two differ. Rotating back to a key whose natural address *is* the account removes the index entry.

```protobuf
message PubKeyHistoryEntry {
  google.protobuf.Any       pub_key             = 1 [(cosmos_proto.accepts_interface) = "cosmos.crypto.PubKey"];
  int64                     replaced_at_height  = 2;
  google.protobuf.Timestamp replaced_at_time    = 3 [(gogoproto.nullable) = false, (gogoproto.stdtime) = true];
  string                    new_key_address     = 4 [(cosmos_proto.scalar) = "cosmos.AddressString"];
}
```

The history is keyed by rotation index rather than block height so that no entry is ever overwritten. Two rotations of the same account in one block, whether in one tx or in separate txs, get distinct keys. A zero-height export restarts block heights, so a height key could collide with an imported entry; an index key cannot. The block height lives only in `replaced_at_height`, which is an `int64` like other block heights in the SDK. Entries for an account are ordered by rotation index, which is also the order in which they happened.

Genesis exports the history per account (`GenesisState.PubKeyHistory`), with each account's entries in rotation-index order. `InitGenesis` assigns indices `0..n-1` in that order, so export and import give identical history. It rebuilds `RekeyIndex` from the accounts rather than reading it from genesis. `ValidateGenesis` requires at most one history per account, each with at least one entry, every history to belong to an account in genesis whose stored pubkey is neither nil nor a `ModuleCredential`, every entry to carry a pubkey that is not a `ModuleCredential` and from which an address can be derived, with a non-negative `replaced_at_height`, the first entry's pubkey to hash to the account address (it is the account's original key), and the entries to form a chain: each entry's `new_key_address` is the natural address of the next entry's pubkey, and the last entry's is the natural address of the account's current pubkey. To keep state consistent with this rule, `RemoveAccount` deletes the account's `PubKeyHistory` along with its `RekeyIndex` pair, so removing an account never leaves history without an account.

The handler preserves the account's concrete type (for example a vesting account stays a vesting account with the same schedule), its account number and its sequence.

### 7. Params (auth consensus version 7 -> 8)

```protobuf
message Params {
  // ... existing fields 1-6 ...
  bool   pub_key_change_enabled = 7;
  uint64 pub_key_change_cost    = 8;
}
```

* `pub_key_change_enabled` defaults to `false`, and the 7 -> 8 migration sets it to `false`. Governance enables it with `MsgUpdateParams`.
* `pub_key_change_cost` defaults to `50000` gas and must be non-zero. The handler charges it on top of normal tx byte and store write gas.

An account that has rekeyed cannot be pruned automatically, because recreating its address would need the original pubkey, which the owner may no longer have. The SDK does not prune accounts today, but `PubKeyChangeCost` compensates for this externality. A future `MsgDeleteAccount` could let rekeyed accounts prune themselves for a gas refund.

```go
ctx.GasMeter().ConsumeGas(params.PubKeyChangeCost, "change pubkey")
```

The handler checks, in order: the param is enabled (`ErrPubKeyChangeDisabled`); the account exists and may rekey under decision 4 (`ErrAccountNotRekeyable`); the new key is acceptable (`ErrInvalidNewPubKey`). It then consumes `PubKeyChangeCost` gas, before the proof is decoded, so a tx whose proof fails still pays for the signature verification it caused. Next it checks that the proof is valid for the doc `{ctx.ChainID(), acc.GetAccountNumber(), msg.Address, msg.NewPubKey}` (`ErrInvalidPubKeyProof`) and runs the registered `PubKeyChangeHooks`. Hooks let modules reject a change that would violate an invariant involving the account's current authentication key. Only after every hook succeeds does the handler append the old key to `PubKeyHistory`, update `RekeyIndex`, set the new pubkey and save the account. It emits a `change_pubkey` event with attributes `address`, `old_pubkey_address` and `new_pubkey_address`.

### 8. Queries

```protobuf
service Query {
  // RekeyedAccounts returns the accounts whose current key's natural address is `address`.
  rpc RekeyedAccounts(QueryRekeyedAccountsRequest) returns (QueryRekeyedAccountsResponse);
  // PubKeyHistory returns the pubkey rotation history of an account.
  rpc PubKeyHistory(QueryPubKeyHistoryRequest) returns (QueryPubKeyHistoryResponse);
}
```

A wallet that recovers from a mnemonic derives the key's natural address and calls `RekeyedAccounts` to find the account the key now controls. `PubKeyHistory` returns the account's history entries.

### 9. Remove address-from-pubkey derivations

Every place that derives an account address from a pubkey must stop doing so:

* `BaseAccount.Validate` drops its `pubkey.Address() == address` check. `types.ValidateGenesis` takes over that role: an account whose stored pubkey is not a `ModuleCredential` and whose `pk.Address() != address` is valid only if `GenesisState.PubKeyHistory` has an entry for that address.
* The default mempool (`types/mempool/signer_extraction_adapter.go`, `types/mempool/sender_nonce.go`) keys senders by `GetSigners()` instead of `sig.PubKey.Address()`. Otherwise a rekeyed account's nonces split across two "senders", and a tx that omits its pubkey dereferences nil.
* Client signing in `client/tx/tx.go` stops deriving `SignerData.Address` from the pubkey (decision 10).
* `x/bank` and `enterprise/group` simulations compare account addresses, rather than pubkeys, when selecting a different account.
* `enterprise/poa` still derives consensus addresses from consensus keys. When enforcing that operator and consensus keys differ, however, it compares the proposed consensus key with both the operator address and the operator account's stored current pubkey. Its auth rekey hook preserves that invariant when an operator changes its authentication key.
* `client/v2/autocli/flag/address.go`, which accepts a pubkey in place of an address, is a client follow-up. `server/start.go` derives validator consensus addresses from consensus keys, which this ADR does not affect.

### 10. Client

A new `--signer-address` flag lets the keyring key named by `--from` sign for a different account address. When it is set, the tx factory uses it for `SignerData.Address` and for the account lookup instead of `pubKey.Address()`. `gentx` also validates the specified account's genesis balance and uses it as the validator operator address. Without the flag, `--from` behaves as before.

An account rekeyed to a multisig spends with the usual multisig flow, passing `--signer-address` to each step: `sign` or `sign-batch` with `--multisig`, then `multisign` or `multisign-batch`. The members' signatures and the combined signature are then made and checked for the rekeyed account, not for the multisig key's own address. `tx validate-signatures` accepts a signature whose pubkey does not hash to its signer when the signer's stored pubkey is that key; in `--offline` mode, where it cannot check, it prints a warning instead. As for any multisig, this flow supports only `SIGN_MODE_LEGACY_AMINO_JSON`: a `SIGN_MODE_DIRECT` sign doc commits to the `AuthInfo`, which differs between what each member signs and the combined tx.

The CLI adds:

* `tx auth sign-rekey-proof [account-address] [new-pubkey-json] --from <key>`, which writes a single signature JSON over the proof doc. Each member of a multisig new key runs it.
* `tx auth change-pubkey [account-address] [new-pubkey-json] --proof <file>[,<file>...] --from <current-key>`, which combines multiple proof files into a `MultiSignatureData`.

### Non-goals

These are separate tracks:

* A "quantum emergency" migration path for accounts whose secp256k1 key is already compromised.
* `client/v2` autocli signing for rekeyed accounts.
* Changing a validator's operator address. It is not needed: the operator address stays fixed and its key becomes rotatable.

## Consequences

### Backwards Compatibility

This is state machine breaking. `x/auth` moves from consensus version 7 to 8 with a params migration, and the ante handler rule changes. The feature is off until governance enables it.

### Positive

* Users and validator operators can rotate keys, which is better operational security.
* Organizations and groups can add and remove multisig signers without moving funds.
* Accounts and validator operators can move from secp256k1 to ML-DSA-65, or to a mixed post-quantum multisig, without moving funds or unbonding.

### Negative

Breaks the current assumed relationship between address and pubkey as H(pubkey) = address. This has a couple of consequences.

* Indexers and wallets can no longer derive an account's address from its pubkey. They must use `RekeyedAccounts` to map a key to the accounts it controls, and `PubKeyHistory` to know which key was active when.
* Wallets that support this feature are more complicated. For example, after a rekey, the CLI must be told the account address with `--signer-address`, because the key's natural address is no longer the account.
* Accounts with zero balance that have rekeyed cannot be pruned automatically.
* Txs signed by the old key before a rotation can sit in the mempool until they fail at inclusion or expire (decision 5).

### Neutral

* While the purpose of this is intended to allow the owner of an account to update to a new pubkey they own, this could technically also be used to transfer ownership of an account to a new owner.  For example, this could be used to sell a staked position without unbonding or an account that has vesting tokens.  However, the friction of this is very high as this would essentially have to be done as a very specific OTC trade. Furthermore, additional constraints could be added to prevent accounts with Vesting tokens to use this feature.
* Will require that PubKeys for an account are included in the genesis exports.

## Test Cases

* A rekeyed account survives genesis export, validation and import, including zero-height export, with identical history and index.
* Two rotations of the same account in one block, in one tx or in two txs, both appear in `PubKeyHistory` and neither overwrites the other.
* An account with a nil stored pubkey that dispatches `MsgChangePubKey` as itself (as a contract or interchain account would) gets `ErrAccountNotRekeyable`, and the account is unchanged.
* A rekeyed account is ordered as a single sender in the mempool.
* A new key that is unsupported, has more than `TxSigLimit` subkeys, or has a threshold the proof can't meet is rejected and the account is unchanged.
* `MsgChangePubKey` reached through authz `MsgExec`, a group proposal or a gov proposal is rejected.
* A tx signed by the old key and submitted after the rotation fails with no state change. If it carries its pubkey, it fails `RecheckTx` too.
* Simulating a tx for a rekeyed account that omits its pubkey succeeds and returns a gas estimate.
* A proof made for account A on chain X is rejected for account B or chain Y.
* In the simulator, the `MsgChangePubKey` operation rotates accounts to secp256k1 or ML-DSA-65 keys, and later operations, including `x/gov` votes scheduled before the rotation, sign with the new key.
* A PoA validator cannot use the same current key for operator authentication and consensus, whether it proposes that key during validator creation or rotation or changes its account key afterward.
* Simulations treat two account addresses using the same current pubkey as distinct accounts.

## References

* https://www.algorand.com/resources/blog/announcing-rekeying
* [ADR-036: Arbitrary Message Signature Specification](./adr-036-arbitrary-signature.md)
* [#26472](https://github.com/cosmos/cosmos-sdk/pull/26472): ML-DSA-65 support for account keys
