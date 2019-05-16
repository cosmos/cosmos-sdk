package genutil

import (
	"encoding/json"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// expected staking keeper
type StakingKeeper interface {
	ApplyAndReturnValidatorSetUpdates(sdk.Context) (updates []abci.ValidatorUpdate)
}

// expected account keeper
type AccountKeeper interface {
	NewAccount(sdk.Context, auth.Account) auth.Account
	SetAccount(sdk.Context, auth.Account)
	IterateAccounts(ctx sdk.Context, process func(auth.Account) (stop bool))
}

// The expected interface for iterating genesis accounts object
type GenesisAccountsIterator interface {
	IterateGenesisAccounts(
		cdc *codec.Codec,
		appGenesis map[string]json.RawMessage,
		iterateFn func(auth.Account) (stop bool),
	)
}
