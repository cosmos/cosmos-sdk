# Gaia 创世状态（Genesis State）

Gaia 创世状态`GenesisState`由账户、各种模块状态和元数据组成，例如创世交易。 每个模块可以指定自己的`GenesisState`。 此外，每个模块可以指定自己的创世状态有效性验证、导入和导出功能。

Gaia 创世状态定义如下：

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

在 Gaia 的 ABCI`initChainer`定义中调用`initFromGenesisState`，它在内部调用每个模块的`InitGenesis`，提供它自己的`GenesisState`作为参数。

## 账户（Accounts）

 `GenesisState` 中的创世账户定义如下：

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

除序列号（nonce）和地址外，每个帐户还必须具有有效且唯一的账户编号。

账户也可能锁仓，此时他们必须提供必要的锁仓信息，锁仓帐户必须至少提供`OriginalVesting`和`EndTime`。如果还提供了`StartTime`，则该帐户将被视为“连续”锁仓帐户，其中按预定义的时间表锁仓 coins。 提供的`StartTime`必须小于`EndTime`，但可能是未来的时间。 换句话说，它不必等于创世时间。 在从新状态（未导出）开始的新链中，`OriginalVesting`必须小于或等于`Coins`。

<!-- TODO: Remaining modules and components in GenesisState -->