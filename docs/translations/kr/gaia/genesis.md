# Gaia 제네시스 스테이트

Gaia의 제네시스 스테이트인 `GenesisState`는 계정 정보, 모듈 스테이트 그리고 제네시스 트랜잭션 같은 메타데이터 등으로 구성됩니다. 각 모듈은 각자의 `GenesisState`를 지정할 수 있습니다. 또한, 각 모듈은 각자의 제네시스 스테이트 검증, 임포트, 엑스포트 기능 등을 지정할 수 있습니다.

Gaia 제네시스 스테이트는 다음과 같이 정의됩니다:

```go
type GenesisState struct {
  Accounts     []GenesisAccount      `json:"accounts"`
  AuthData     auth.GenesisState     `json:"auth"`
  BankData     bank.GenesisState     `json:"bank"`
  StakingData  staking.GenesisState  `json:"staking"`
  MintData     mint.GenesisState     `json:"mint"`
  DistrData    distr.GenesisState    `json:"distr"`
  GovData      gov.GenesisState      `json:"gov"`
  SlashingData slashing.GenesisState `json:"slashing"`
  GenTxs       []json.RawMessage     `json:"gentxs"`
}
```

ABCI `initChainer`에서는 Gaia의 `initFromGenesisState`를 기반으로 각 모듈의 `InitGenesis`를 호출해 각 모듈들의 `GenesisState`를 파라미터 값으로 불러옵니다.

## 계정

`GenesisState`에서 제네시스 계정은 다음과 같이 정의됩니다:

```go
type GenesisAccount struct {
  Address       sdk.AccAddress `json:"address"`
  Coins         sdk.Coins      `json:"coins"`
  Sequence      uint64         `json:"sequence_number"`
  AccountNumber uint64         `json:"account_number"`

  // vesting account fields
  OriginalVesting  sdk.Coins `json:"original_vesting"`  // total vesting coins upon initialization
  DelegatedFree    sdk.Coins `json:"delegated_free"`    // delegated vested coins at time of delegation
  DelegatedVesting sdk.Coins `json:"delegated_vesting"` // delegated vesting coins at time of delegation
  StartTime        int64     `json:"start_time"`        // vesting start time (UNIX Epoch time)
  EndTime          int64     `json:"end_time"`          // vesting end time (UNIX Epoch time)
}
```

각 계정은 시퀀스 수(sequence number (nonce))와 주소 외에도 유효한 고유 계정 번호를 보유해야 합니다.

만약 계정이 베스팅 계정인 경우, 필수 베스팅 정보가 제공되어야 합니다. 베스팅 계정은 최소 `OriginalVestin` 값과 `EndTime` 값이 정의되어야 합니다. 먄약 `StartTime`이 함께 정의된 경우, 계정은 "연속되는(continuous)" 베스팅 계정으로 처리되며, 지정된 스케줄 안에서 꾸준히 토큰을 언락합니다. 여기에서 `StartTime`의 값은 `EndTime`의 값 보다 작아야 하지만, `StartTime`의 값은 미래 값으로 지정할 수는 있습니다 (제네시스 시간과 동일하지 않아도 괜찮습니다). 새로운 스테이트(엑스포트 되지 않은 스테이트)에서 시작하는 체인의 경우, `OriginalVestin`의 값은 `Coins`의 값과 동일하거나 적어야 합니다.

<!-- TODO: Remaining modules and components in GenesisState -->