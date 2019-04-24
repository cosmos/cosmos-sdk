# Gaia Genesis State

Gaia genesis state, `GenesisState`, is composed of accounts, various module
states and metadata such as genesis transactions. Each module may specify its
own `GenesisState`. In addition, each module may specify its own genesis state
validation, import and export functionality.

The Gaia genesis state is defined as follows:

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

In the ABCI `initChainer` definition of Gaia the `initFromGenesisState` is called
which internally calls each module's `InitGenesis` providing its own respective
`GenesisState` as a parameter.

## Accounts

Genesis accounts defined in the `GenesisState` are defined as follows:

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

Each account must have a valid and unique account number in addition to a
sequence number (nonce) and address.

Accounts may also be vesting, in which case they must provide the necessary vesting
information. Vesting accounts must provide at a minimum `OriginalVesting` and
`EndTime`. If `StartTime` is also provided, the account will be treated as a
"continuous" vesting account in which it vests coins at a predefined schedule.
Providing a `StartTime` must be less than `EndTime` but may be in the future.
In other words, it does not have to be equal to the genesis time. In a new chain
starting from a fresh state (not exported), `OriginalVesting` must be less than
or equal to `Coins.`

<!-- TODO: Remaining modules and components in GenesisState -->
