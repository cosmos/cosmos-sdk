package std

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/KiraCore/cosmos-sdk/baseapp"
	paramstypes "github.com/KiraCore/cosmos-sdk/x/params/types"
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
			baseapp.ParamStoreKeyEvidenceParams, abci.EvidenceParams{}, baseapp.ValidateEvidenceParams,
		),
		paramstypes.NewParamSetPair(
			baseapp.ParamStoreKeyValidatorParams, abci.ValidatorParams{}, baseapp.ValidateValidatorParams,
		),
	)
}
