# ADR 011: Generalize Genesis Accounts

## Changelog

- 2019-08-30: initial draft

## Context

Summary: The `auth` module allows custom account types, but the `genaccounts` module does not.

Currently the SDK allows for custom account types; the `auth` keeper stores any type fulfilling its Account interface. However `auth` does not handle exporting or loading accounts to/from a genesis file, this is done by `genaccounts`, which only handles one of 4 concrete account types (base, continuousVesting, delayedVesting, module).

Projects desiring to use custom accounts (say custom vesting accounts) need to fork and modify `genaccounts`.


## Decision

In summary, we will (Un)Marshal all accounts (interface types) directly using amino, rather than converting to genaccount’s `GenesisAccount` type. Since doing this removes the majority of genaccount’s code, we will merge genaccounts into auth. Marshalled accounts will be stored in auth's genesis state.

Detailed changes:

#### 1) (un)marshal accounts directly using amino

`auth.GenesisState` gains a new field `Accounts`
```go
// GenesisState - all auth state that must be provided at genesis
type GenesisState struct {
	Params   Params             `json:"params" yaml:"params"`
	Accounts []exported.Account `json:"accounts" yaml:"accounts"`
}
```
`auth.InitGenesis` and `auth.ExportGenesis` (un)marshal accounts as well as params.
```go
// InitGenesis - Init store state from genesis data
func InitGenesis(ctx sdk.Context, ak AccountKeeper, data GenesisState) {
	ak.SetParams(ctx, data.Params)
	// load the accounts
	for _, a := range data.Accounts {
		acc := ak.NewAccount(ctx, a) // set account number
		ak.SetAccount(ctx, acc)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper
func ExportGenesis(ctx sdk.Context, ak AccountKeeper) GenesisState {
	params := ak.GetParams(ctx)
	accounts := ak.GetAllAccounts(ctx)
	return NewGenesisState(params, accounts)
}
```

#### 2) registering custom account types on the auth codec

The auth codec must have all custom account types registered to marshal them. We will follow the pattern established in gov for proposals.

A custom account definition:
```go
import authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
// Register the module account type with the auth module codec so it can decode module accounts stored in a genesis file
func init() {
	authtypes.RegisterAccountTypeCodec(ModuleAccount{}, "cosmos-sdk/ModuleAccount")
}
type ModuleAccount struct {
    ...
```
The auth codec definition:
```go
var ModuleCdc *codec.Codec
func init() {
    ModuleCdc = codec.New()
    // register module msg's and Account interface
    ...
	// leave the codec unsealed
}
// RegisterAccountTypeCodec registers an external account type defined in another module for the internal ModuleCdc.
func RegisterAccountTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}
```

#### 3) genesis validation for custom account types
Modules implement a `ValidateGenesis` method. As auth does not know of account implementations, accounts will need to validate themselves.

We will add a `Validate` method to the `auth.Account` interface.

Then auth.ValidateGenesis becomes:
```go
// ValidateGenesis performs basic validation of auth genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	// Validate params
	...

	// Validate accounts
	addrMap := make(map[string]bool, len(data.Accounts))
	for _, acc := range data.Accounts {

		// check for duplicated accounts
		addrStr := acc.GetAddress().String()
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate account found in genesis state; address: %s", addrStr)
		}
		addrMap[addrStr] = true

		// check account specific validation
		if err := acc.Validate(); err != nil {
			return fmt.Errorf("invalid account found in genesis state; address: %s, error: %s", addrStr, err.Error())
		}

	}
	return nil
}
```

#### 4) move add-genesis-account cli to auth
Genaccounts contains a cli command to add base or vesting accounts to a genesis file.

This will be moved to auth. We will leave it to projects to write their own command to add custom accounts. An extensible cli handler, similar to gov, could be created but it is not worth the complexity for this minor use case.

#### 5) update module and vesting accounts
Under the new scheme module and vesting account types need some minor updates:

 - type registration on auth (shown above)
 - A `Validate` method


## Status

Proposed

## Consequences

### Positive

 - custom accounts can be used without needing to fork `genaccounts`
 - reduction in lines of code

### Negative

- expansion of Account interface

### Neutral

- genaccounts module no longer exists
- accounts in genesis files are stored under `auth.accounts` rather than `accounts`
- `add-genesis-account` cli command now in auth

## References
