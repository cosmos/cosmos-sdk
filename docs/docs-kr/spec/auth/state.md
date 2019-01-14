## 상태(State)

### 계정

계정은 SDK 블록체인 내에서 신원이 확보된 고유 유저의 인증 정보를 포함하고 있습니다. 이 정보에는 리플레이 방지를 위한 퍼블릭키, 주소, 계정 번호 / 순서 번호(sequence number) 등이 포함되어 있습니다. 수수료를 지불하기 위해서 계정 잔고를 조회해야 하기 때문에, 효율성을 위해서 특정 계정의 잔고는 `sdk.Coins`로 저장합니다.

계정은 외부적으로(externally) 인터페이스로 노출이 되며, 내부적으로는 base account 또는 vesting account의 형태로 저장됩니다. 추가적인 계정 형태를 추가하는 것을 원하는 모듈 클라이언트는 별도로 추가가 가능합니다.

- `0x01 | Address -> amino(account)`

#### 계정 인터페이스

계정 인터페이스는 기본 계정 정보를 읽고 쓰는 메소드(method)를 정의합니다. 참고로 모든 메소드는 인터페이스를 준수하는 구조로 작동합니다. 특정 계정을 스토어에 쓰기 위해서는 게정 키퍼(keeper)가 사용되어야 합니다.

```golang
type Account interface {
  GetAddress() AccAddress
  SetAddress(AccAddress)

  GetPubKey() PubKey
  SetPubKey(PubKey)

  GetAccountNumber() uint64
  SetAccountNumber(uint64)

  GetSequence() uint64
  SetSequence(uint64)

  GetCoins() Coins
  SetCoins(Coins)
}
```

#### Base Account

베이스 어카운트는 가장 흔하고 기본적인 계정 형태입니다. 베이스 어카운트는 모든 값을 struct 형태로 직접적으로 저장합니다.


```golang
type BaseAccount struct {
  Address       AccAddress
  Coins         Coins
  PubKey        PubKey
  AccountNumber uint64
  Sequence      uint64
}
```

#### Vesting Account

[Vesting](vesting.md)을 확인하세요.
