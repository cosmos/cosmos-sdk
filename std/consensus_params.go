package std

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ConsensusParamsKeyTable returns an x/params module keyTable to be used in
// the BaseApp's ParamStore. The KeyTable registers the types along with the
// standard validation functions. Applications can choose to adopt this KeyTable
// or provider their own when the existing validation functions do not suite their
// needs.
func ConsensusParamsKeyTable() paramstypes.KeyTable {
	return paramstypes.NewKeyTable(
		paramstypes.NewParamSetPair(
			baseapp.ParamStoreKeyBlockParams, abci.BlockParams{}, baseapp.ValidateBlockParams,
		),
		paramstypes.NewParamSetPair(
			baseapp.ParamStoreKeyEvidenceParams, tmproto.EvidenceParams{}, baseapp.ValidateEvidenceParams,
		),
		paramstypes.NewParamSetPair(
			baseapp.ParamStoreKeyValidatorParams, tmproto.ValidatorParams{}, baseapp.ValidateValidatorParams,
		),
	)
}
