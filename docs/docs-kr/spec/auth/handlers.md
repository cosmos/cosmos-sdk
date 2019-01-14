## 핸들러

현재 상태로 auth 모듈은 자체적인 트랜잭션 핸들러가 없습니다. 다만, 특수 `AnteHandler`를 이용하여 기본적인 트랜잭션의 유효성 체크를 실행하며 멤풀에서 버릴 트랜잭션을 필터링 합니다. 참고로 현재 텐더민트 제안자는 `CheckTx`를 실패하는 트랜잭션도 블록에 추가할 수 있다는 특수성이 아직 존재하기 때문에 안티핸들러는 `CheckTx` 뿐만 아니라 `DeliverTx`에서도 이용된다는 점을 참고하시기 바랍니다. 

### 안티핸들러(Ante Handler)

```golang
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
