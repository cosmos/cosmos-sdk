# Validator Consensus Key Rotation

`MsgRotateConsPubKey` lets a validator replace its consensus public key without
being removed and recreated. The rotation is applied in the same block the
transaction lands. Power, metadata, and accrued fees are preserved.

## Message

```proto
message MsgRotateConsPubKey {
  option (cosmos.msg.v1.signer) = "sender";

  string              sender            = 1;
  string              validator_address = 2;
  google.protobuf.Any new_pub_key       = 3;
}
```

| Field | Description |
|-------|-------------|
| `sender` | The transaction signer. Must be the validator operator or the module admin. |
| `validator_address` | Operator address of the validator whose key is rotated. This is the stable identifier; it does not change. |
| `new_pub_key` | The new consensus public key, packed as an `Any`. Its type must be valid for the chain's consensus params. |

### Authorization

The signer is `sender`. The handler accepts the message when:

- `sender == validator_address` (operator rotates its own key), or
- `sender == Params.Admin` (admin overrides for any validator).

Any other sender is rejected with `ErrUnauthorized`.

### What it does

1. Looks up the validator by operator address.
2. Validates the new key: correct type, not equal to the current key (no-op is
   rejected), not already in use by another validator, and the operator address
   must not derive from the new consensus key.
3. Checkpoints pending fees under the old consensus address, then re-keys the
   record from the old consensus address to the one derived from the new key.
   Power, metadata, and operator are preserved, so `totalPower` is unchanged.
4. Migrates the accrued fee entry from the old consensus address to the new one.
5. Queues ABCI validator updates when the validator has power (see Power-0
   below).
6. Emits `EventTypeRotateConsPubKey` with the operator, old consensus address,
   and new consensus address.

There is no rotation history. The old key is simply gone after the swap.

## Design rationale

**No fee, no rate limit.** The POS key-rotation fee and rate limit exist to make
evidence-based slashing tractable: a rotated-away key must remain attributable
for the unbonding window, and the fee discourages churn that complicates that
bookkeeping. POA has no slashing and no evidence handling, so neither concern
applies. Rotation is free and unthrottled.

**Auto-migrate accrued fees.** A validator's allocated fees are stored per
consensus address. Rotation moves that entry from the old address to the new one
so the operator's balance follows the key. The total allocated is untouched
because the balance moves rather than changes. There is no forced payout; the
operator withdraws on its own schedule as before, and the migration is invisible
from the operator's point of view.

**Same-block swap, no deferred queue.** Because there is no evidence window to
keep the old key attributable, the state swap happens immediately in the block
that includes the transaction. There is no rotation history and no apply queue
on the SDK side.

**Power-0 validators emit no ABCI update.** A validator with power 0 is not in
CometBFT's active set. An ABCI update with power 0 is treated by CometBFT as a
deletion, so emitting `old@0` and `new@0` would remove addresses that are not in
the set and abort the block. A power-0 rotation therefore re-keys internal state
only and queues no ABCI updates. When the validator has power, the rotation
queues `old@0` (removes the old key from the active set) and `new@power` (adds
the new key).

## Operator runbook

The SDK swaps validator state in the same block as the rotation transaction, but
CometBFT does not start expecting the new key to sign immediately. The
`old@0` + `new@power` updates are returned at `EndBlock` and CometBFT applies the
validator-set change on its own block delay, a couple of blocks later. Until that
change applies, CometBFT still expects the **old** consensus key to sign. The
operational risk is in this window:

- Swap the node's `priv_validator_key.json` to the new key too early and the node
  signs with a key CometBFT does not yet expect, missing blocks (downtime).
- Run two signers at once, one with the old key and one with the new, and the
  validator can double-sign.

Never have more than one node holding this validator's signing key active at a
time.

Ordered steps:

1. Generate the new consensus key on a secondary node or offline. Keep the
   current node running and signing with the **old** key. Do not touch the live
   `priv_validator_key.json` yet.
2. Extract the new consensus public key (base64) and its type, for example with
   `simd comet show-validator` against the new key material.
3. Submit the rotation (signer is the operator):

   ```bash
   simd tx poa rotate-cons-pub-key <new_pubkey_base64> ed25519 \
       --operator-address $OPERATOR_ADDR \
       --from myvalidator \
       --keyring-backend test
   ```

   The transaction re-keys state and migrates fees in the same block.
4. Watch the active validator set (`simd q poa validators`) and consensus until
   the new consensus address appears and the old one is gone. This is when
   CometBFT has applied the update and now expects the **new** key to sign.
5. **Dangerous step.** At that point, and only then, stop the node, replace
   `priv_validator_key.json` with the new key, and restart. Confirm the node is
   signing under the new consensus address.

If you run a redundant standby, the safe pattern is: keep the old-key node
signing until the update applies, cut over to the new-key node in a single
switch, and confirm the old-key node is fully stopped before the new-key node
starts signing. Both signing at once risks a double-sign.

**Power-0 validators:** there is no active-set transition to time, so no ABCI
update is emitted. Make sure the node uses the new key before the admin gives the
validator power.

## Admin runbook

The admin can rotate any validator's key by setting `--from` to the admin key.
The command and arguments are identical; only the signer differs. Use the admin
override when:

- **Unresponsive operator.** The operator cannot or will not rotate, and the key
  needs to change (for example, a planned infrastructure migration).
- **Suspected key compromise.** Rotate immediately to a fresh key the operator
  controls, cutting off a leaked consensus key.
- **Recovery.** Restore a validator to a known-good key the operator can sign
  with.

```bash
simd tx poa rotate-cons-pub-key <new_pubkey_base64> ed25519 \
    --operator-address $OPERATOR_ADDR \
    --from admin \
    --keyring-backend test
```

The admin override only changes on-chain state. The operator (or whoever runs the
node) must still swap `priv_validator_key.json` per the operator runbook timing,
otherwise the validator goes dark until its node signs with the new key. Coordinate
the node-side swap with the operator before submitting, unless the intent is
precisely to revoke a compromised key and the corresponding node should stop
signing.
