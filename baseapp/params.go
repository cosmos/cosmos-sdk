package baseapp

import (
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Paramspace defines the parameter subspace to be used for the paramstore.
const Paramspace = "baseapp"

// Parameter store keys for all the consensus parameter types.
var (
	ParamStoreKeyBlockParams     = []byte("BlockParams")
	ParamStoreKeyEvidenceParams  = []byte("EvidenceParams")
	ParamStoreKeyValidatorParams = []byte("ValidatorParams")
)

// ParamStore defines the interface the parameter store used by the BaseApp must
// fulfill.
type ParamStore interface {
	Get(ctx sdk.Context, key []byte) (tmproto.ConsensusParams, error)
	Has(ctx sdk.Context, key []byte) bool
	Set(ctx sdk.Context, cp tmproto.ConsensusParams)
}
