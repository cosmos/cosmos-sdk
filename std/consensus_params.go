package std

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// ConsensusParamsKeyTable returns an x/params module keyTable to be used in
// the BaseApp's ParamStore. The KeyTable registers the types along with the
// standard validation functions. Applications can choose to adopt this KeyTable
// or provider their own when the existing validation functions do not suite their
// needs.
func ConsensusParamsKeyTable() params.KeyTable {
	return params.NewKeyTable(
		params.NewParamSetPair(
			baseapp.ParamStoreKeyBlockParams, abci.BlockParams{}, baseapp.ValidateBlockParams,
		),
		params.NewParamSetPair(
			baseapp.ParamStoreKeyEvidenceParams, abci.EvidenceParams{}, baseapp.ValidateEvidenceParams,
		),
		params.NewParamSetPair(
			baseapp.ParamStoreKeyValidatorParams, abci.ValidatorParams{}, baseapp.ValidateValidatorParams,
		),
	)
}
