<!--
order: 3
-->

# AnthHandlers

## Handlers

The auth module presently has no transaction handlers of its own, but does expose
the special `AnteHandler`, used for performing basic validity checks on a transaction,
such that it could be thrown out of the mempool. Note that the ante handler is called on
`CheckTx`, but _also_ on `DeliverTx`, as Tendermint proposers presently have the ability
to include in their proposed block transactions which fail `CheckTx`.

### Ante Handler

```go
anteHandler(ak AccountKeeper, fck FeeCollectionKeeper, tx sdk.Tx)
  if !tx.(StdTx)
    fail with "not a StdTx"

  if isCheckTx and tx.Fee < config.SubjectiveMinimumFee
    fail with "insufficient fee for mempool inclusion"

  if tx.ValidateBasic() != nil
    fail with "tx failed ValidateBasic"

  if tx.Fee > 0
    account = GetAccount(tx.GetSigners()[0])
    coins := acount.GetCoins()
    if coins < tx.Fee
      fail with "insufficient fee to pay for transaction"
    account.SetCoins(coins - tx.Fee)
    fck.AddCollectedFees(tx.Fee)

  for index, signature in tx.GetSignatures()
    account = GetAccount(tx.GetSigners()[index])
    bytesToSign := StdSignBytes(chainID, acc.GetAccountNumber(),
      acc.GetSequence(), tx.Fee, tx.Msgs, tx.Memo)
    if !signature.Verify(bytesToSign)
      fail with "invalid signature"

  return
```
