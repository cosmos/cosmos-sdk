package simulation

import (
	"math/rand"

	"cosmossdk.io/x/auth/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// GenerateMsgUpdateParams returns a random MsgUpdateParams.
func GenerateMsgUpdateParams(r *rand.Rand) sdk.Msg {
	// use the default gov module account address as authority
	var authority sdk.AccAddress = address.Module("gov")

	params := types.DefaultParams()
	params.MaxMemoCharacters = uint64(simtypes.RandIntBetween(r, 1, 1000))
	params.TxSigLimit = uint64(simtypes.RandIntBetween(r, 1, 1000))
	params.TxSizeCostPerByte = uint64(simtypes.RandIntBetween(r, 1, 1000))
	params.SigVerifyCostED25519 = uint64(simtypes.RandIntBetween(r, 1, 1000))
	params.SigVerifyCostSecp256k1 = uint64(simtypes.RandIntBetween(r, 1, 1000))

	return &types.MsgUpdateParams{
		Authority: authority.String(),
		Params:    params,
	}
}
