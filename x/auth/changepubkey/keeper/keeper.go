package keeper

import (
	"fmt"
	"time"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

// Keeper manages history of public keys per account
type Keeper struct {
	key sdk.StoreKey
	cdc codec.BinaryMarshaler
}

// NewKeeper returns a new keeper which manages pubkey history per account.
func NewKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey,
) Keeper {

	return Keeper{
		key: key,
		cdc: cdc,
	}
}

// "Everytime a key for an address is changed, we will store a log of this change in the state of the chain,
// thus creating a stack of all previous keys for an address and the time intervals for which they were active.
// This allows dapps and clients to easily query past keys for an account which may be useful for features
// such as verifying timestamped off-chain signed messages."

// Logger returns a module-specific logger.
func (pk Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetPubKeyHistory Returns the PubKey history of the account at address by time
func (pk Keeper) GetPubKeyHistory(ctx sdk.Context, addr sdk.AccAddress) ([]types.PubKeyHistory, error) {
	// TODO: implement full history getter function
	return []types.PubKeyHistory{}, nil
}

// GetPubKeyHistoricalEntry Returns the PubKey historical entry at a specific time
func (pk Keeper) GetPubKeyHistoricalEntry(ctx sdk.Context, addr sdk.AccAddress, time time.Time) (types.PubKeyHistory, error) {
	// TODO: implement account / time query
	return types.PubKeyHistory{}, nil
}

// StoreLastPubKey Store pubkey of an account at the time of changepubkey action
func (pk Keeper) StoreLastPubKey(ctx sdk.Context, addr sdk.AccAddress, time time.Time) error {
	// TODO: implement previous pubkey set action
	return nil
}
