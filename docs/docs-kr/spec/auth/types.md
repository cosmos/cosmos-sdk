## 종류(Types)

[상태](state.md)에 정의된 계정 외에도 auth 모듈은 `StdFee`(수수료 수량 + 가스 리밋의 조합), `StdSignature`(선택사항(optional)인 퍼블릭 키와 암호화 서명을 바이트 어레이로 표현), `StdTx`(`sdk.tx` 인터페이스를 `StdFee`와 `StdSignature`를 구조화함), 그리고 `StdSignDoc`(트랜잭션을 보내는 유저가 사인해야 하는 리플레이 방지 기능)을 사용합니다.

### StdFee

`StdFee`는 수수료의 수량(아무 토큰 단위는 가능)과 가스 리밋(`수량/가스 리밋 = 가스 비용`)의 조합입니다. 

```golang
type StdFee struct {
  Amount Coins
  Gas    uint64
}
```

### StdSignature

`StdSignature`는 퍼블릭 키와 암호화 서명을 바이트 어레이(byte array)로 구조화한 것입니다. SDK는 특정 키 포맷과 서명 포맷을 표준화 하지 않으며 `PubKey` 인터페이스가 지원하는 모든 데이터를 지원합니다.

```golang
type StdSignature struct {
  PubKey    PubKey
  Signature []byte
}
```

### StdTx

`StdTx`는 `sdk.Tx`를 구조화한 것입니다. 이 구조는 매우 기본적이게 설계되었기 때문에 대다수의 코스모스 SDK 블록체인과 호환될 수 있을 것으로 예상하고 있습니다.

```golang
type StdTx struct {
  Msgs        []sdk.Msg
  Fee         StdFee  
  Signatures  []StdSignature
  Memo        string
}
```

### StdSignDoc

`StdSignDoc`은 트랜잭션을 전송하는 유저가 동일한 블록체인에서 동일한 트랜잭션을 다중으로 전파하는 리플레이 방지(replay-protection) 시스템입니다. 유저는 `StdSignDoc`를 서명을 하기 때문에 특정 블록체인에서 다중 트랜잭션을 발생시킬 수 없습니다.

미래 호환성을 고려해 SDK types가 아닌 `json.RawMessage`로 작성되는 것을 장려합니다.

```golang
type StdSignDoc struct {
  AccountNumber uint64
  ChainID       string
  Fee           json.RawMessage
  Memo          string
  Msgs          []json.RawMessage
  Sequence      uint64
}
```
