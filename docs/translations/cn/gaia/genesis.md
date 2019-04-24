# Gaia的创世状态

`GenesisState`是Gaia的创世状态，由账户，不同模块的状态和例如创世交易这样的数据组成。每个模块可以指定自己的`GenesisState`，还有每个模块可以指定对创世状态的验证，导入和导出。

Gaia的创世状态有如下定义:

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

在gaia的ABCI接口`initChainer`的定义中`initFromGenesisState`被调用，它在内部调用每个模块的`InitGenesis`，提供各自的`GenesisState`作为参数

## 账户

`GenesisState`中的创世账户有如下定义:

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

每个账户必须有一个合理的并唯一的account number，还要有sequence number和address。

账户也可以是锁定账户，必须提供必要的锁定信息。锁定账户必须提供一个最小的`OriginalVesting`和`EndTime`。如果`StartTime`也提供了，这个账户将会被当做一个连续的锁定账户，将按照预定的时间线锁定代币。提供的`StartTime`必须小于`EndTime`，但可以是将来的某个时间。换句话说，`StartTime`不必小于创世时间。当一条新链从一个新状态（不是到处的）生成时，`OriginalVesting` 必须要小于`Coins`

<!-- TODO: Remaining modules and components in GenesisState -->

