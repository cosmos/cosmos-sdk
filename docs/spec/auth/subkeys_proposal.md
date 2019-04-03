# Subkeys Proposal Spec

## Abstract

Currently, a `StdAccount` has only one public key assoiciated with it, and all transactions that are made with that key must be signed by that one public key. This is problematic because the security level required by different types of transactions executable by an account may be different. For example, a user may want to keep the key that has the ability to send their funds in cold storage, but keep a key that can vote on governance proposals on their phone.

For this reason, we introduce the concept of SubKeys.  With this proposal, accounts will be able to have multiple public keys, each of which can have different permissions.  The main public key of an account can provision new subkeys with specific msg routes which the subkey is allowed to sign for.  If a subkey tries to sign a tx with a msg type it is not permitted to use, the tx will be rejected.

One complexity in this system is that all txs, regardless of their msg type, need to be able to send tx fees. This means that even your lowest security subkey, if compromised, could drain your account by using your entire balance as the tx fee for a transaction. To resolve this, we add a notion of FeeAllowances.  The master pubkey, along with provisioning subkeys with msg types they are allowed to use, also provisions them with a daily allowance of tokens that they are allowed to spend to pay transaction fees.  For example, if SubKey A has a daily fee allowance of 1atom, it can spend up to 1atom on transaction fees in a 24 hour window.  If it does not use its atom in a 24 hour window, the "allowance" does not accumulate, it still stays at 1atom in any 24 hour window.


## Construction

To begin, we define the struct that stores the metadata about a certain account.  This includes the pubkey itself, the routes it's permitted to sign transactions for, its daily allowance, and how much of its allowance has been used in the last 24 hour window.

```golang
type SubKeyMetadata struct {
  PubKey               sdk.AccPubKey
  PermissionedRoutes   []string
  DailyFeeAllowance    sdk.Coins
  DailyFeeUsed         sdk.Coins
  Revoked              bool
}
```

We create a new `Account` type called `SubKeyAccount` that stores a list of `SubKeyMetadata`s.

```golang
type SubKeyAccount struct {
  Address       AccAddress
  Coins         Coins
  PubKey        sdk.AccPubKey
  AccountNumber uint64
  Sequence      uint64
  SubKeys       []SubKeyMetadata
}
```

In the `StdSignature` type we need to add a new field called `PubKeyIndex` which tells the anteHandler which SubKey to use to verify this signature.

```golang
// StdSignature represents a sig
type StdSignature struct {
  PubKey        crypto.PubKey
  Signature     []byte
  PubKeyIndex   uint
}
```

If a `StdSignature.PubKeyIndex` is 0, this means that the account's "master" PubKey (the `SubKeyAccount.PubKey`) should be used to verify the signature.  If the `StdSignature.PubKeyIndex` is `> 0`, it verifies the signature using the corresponding **1-indexed** SubKey in the `SubKeyAccount.SubKeys` slice.  Note, the SubKeys slice is 1-indexed rather than 0-indexed because the 0 index is used to refer to the master PubKey.  See pseudocode below:

```
if stdSig.PubKeyIndex == 0 {
  verify stdSig using acc.PubKey
} else {
  verify stdSig using acc.SubKeys[stdSig.PubKeyIndex - 1]
}
```

In the AnteHandler, if a transaction comes from a SubKey, it verifies that the SubKey has not been revoked and it is permitted to use the msg routes it is trying to sign over.

```
if acc.SubKeys[stdSig.PubKeyIndex - 1].Revoked {
  return ErrNotPermitted
}
if msg.Route not in subKeyMetadata.PermissionedRoutes {
  return ErrNotPermitted
}
```

When a transaction signed by a SubKey is accepted and has paid fees, we need to log the fees it has paid.  We increase the `SubKeyMetadata.DailyFeeUsed` field.

If a new transaction from a subkey comes in and the `tx.Fee + subKeyMetadata.DailyFeeUsed > subKeyMetadata.DailyFeeAllowance`, then the transaction is rejected as the subkey is trying to exceed its daily fee allowance.

But now we need a way to decrease the `DailyFeeUsed` field once transactions are past the 24 hour window.  To do this we us a time based iterator similar to the ones used in the governance and staking queues.

We start by defining a new struct called a `DailyFeeSpend`, which is what will be inserted into the Queue.

```golang
type DailyFeeSpend struct {
  Address              sdk.AccAddress
  SubKeyIndex          uint
  FeeSpent             sdk.Coins
}
```

When a transaction using a SubKey has paid its fee, we insert a new entry into this queue's store.
- The key is `sdk.FormatTimeBytes(ctx.BlockTime.Add(time.Day * 1)`
- The value is the Amino marshalled `DailyFeeSpend`

Finally in the EndBlocker we iterate over the all the DailyFeeSpends in the queue which are expired (more than a day old).  We prune them from state, and deduct their FeeSpent amount from the corresponding SubKeyMetadata.

```
store := ctx.Store(spentFeeQueueStoreKey)
iterator := sdk.Iterator(nil, sdk.FormatTimeBytes(ctx.BlockTime)) // pseudocode abstract this as a slice of DailyFeeSpend

for _, dailyFeeSpend := range iterator {
  acc := accountKeeper.GetAccount(dailyFeeSpend.Address)
  acc.SubKeys[dailyFeeSpend.SubKeyIndex - 1].DailyFeeUsed -= dailyFeeSpend.FeeSpent
  store.Delete(dailyFeeSpend)
}
```

## Msgs

In order to add, revoke, and update SubKeys, we would need to add some new `Msg` types to the auth module.

We start by defining an `MsgAddSubKey` struct:

```golang
type MsgAddSubKey struct {
  Address              sdk.AccAddress
  PubKey               sdk.AccPubKey
  PermissionedRoutes   []string
  DailyFeeAllowance    sdk.Coins
}
```

The handler of this Msg will create a new SubKeyMetadata struct using the data from this Msg, and append it to the end of the SubKeys slice of the Account associated with the Address.

```
acc := accountKeeper.GetAccount(msg.Address)

acc.SubKeys := append(acc.SubKeys, SubKeyMetadata{
  PubKey               msg.PubKey
  PermissionedRoutes   msg.PermissionedRoutes
  DailyFeeAllowance    msg.DailyFeeAllowance
  DailyFeeUsed         sdk.Coins{}
  Revoked              false
})

accountKeeper.SetAccount(acc)
```

To make it easier to reason about the security of a SubKey, we do not allow you to add or remove permissions from a SubKey. If you want a SubKey with new permissions, you should create a new SubKey.  We do however allow you to update the allowance of a specific subkey. This can be useful because minimum transaction fee costs may fluctuate over time.

To update a SubKey, we create a new `MsgUpdateSubKeyAllowance`

```golang
type MsgMsgUpdateSubKeyAllowance struct {
  Address              sdk.AccAddress
  SubKeyIndex          uint
  DailyFeeAllowance    sdk.Coins
}
```

The handler of this message will update the SubKey at the index of `SubKeyIndex` in the account for `Address`.

```
if SubKeyIndex == 0 {
  return Err
}

acc := accountKeeper.GetAccount(msg.Address)

acc.SubKeys[msg.SubKeyIndex - 1].DailyFeeAllowance = msg.DailyFeeAllowance

accountKeeper.SetAccount(acc)
```

Finally, to revoke SubKeys, we add a `MsgRevokeSubKey`

```golang
type MsgMsgUpdateSubKeyAllowance struct {
  Address              sdk.AccAddress
  SubKeyIndex          uint
}
```

The handler of this message will revoke the SubKey at the index of `SubKeyIndex` in the account for `Address`.

```
if SubKeyIndex == 0 {
  return Err
}

acc := accountKeeper.GetAccount(msg.Address)

acc.SubKeys[msg.SubKeyIndex - 1].Revoked = true

accountKeeper.SetAccount(acc)
```

We do not delete the SubKey from state, because removing it from state would mess up the indexing of subkeys.
We also do not provide a mechanism to unrevoke a SubKey.  If you wish to unrevoke a specific SubKey, you can
recreate a SubKey with the same `crypto.PubKey` and permissions at a new SubKey index.
