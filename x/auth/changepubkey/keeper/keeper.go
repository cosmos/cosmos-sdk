package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ChangePubKeyKeeper encodes/decodes accounts using the go-amino (binary)
// encoding/decoding library.
type ChangePubKeyKeeper struct {
	key           sdk.StoreKey
	cdc           codec.BinaryMarshaler
	paramSubspace paramtypes.Subspace
}

// NewChangePubKeyKeeper returns a new sdk.ChangePubKeyKeeper that uses go-amino to
// (binary) encode and decode concrete sdk.Accounts.
func NewChangePubKeyKeeper(
	cdc codec.BinaryMarshaler, key sdk.StoreKey, paramstore paramtypes.Subspace,
) ChangePubKeyKeeper {

	// set KeyTable if it has not already been set
	if !paramstore.HasKeyTable() {
		paramstore = paramstore.WithKeyTable(types.ParamKeyTable())
	}

	return ChangePubKeyKeeper{
		key:           key,
		cdc:           cdc,
		paramSubspace: paramstore,
	}
}
