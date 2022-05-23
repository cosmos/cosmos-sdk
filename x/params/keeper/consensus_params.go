package keeper

import (
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/params/types"
)

// ConsensusParamsKeyTable returns an x/params module keyTable to be used in
// the BaseApp's ParamStore. The KeyTable registers the types along with the
// standard validation functions. Applications can choose to adopt this KeyTable
// or provider their own when the existing validation functions do not suite their
// needs.
func ConsensusParamsKeyTable() types.KeyTable {
	return types.NewKeyTable(
		types.NewParamSetPair(
			baseapp.ParamStoreKeyBlockParams, abci.BlockParams{}, baseapp.ValidateBlockParams,
		),
		types.NewParamSetPair(
			baseapp.ParamStoreKeyEvidenceParams, tmproto.EvidenceParams{}, baseapp.ValidateEvidenceParams,
		),
		types.NewParamSetPair(
			baseapp.ParamStoreKeyValidatorParams, tmproto.ValidatorParams{}, baseapp.ValidateValidatorParams,
		),
		// This param is stored in the param store in order to send it to Tendermint after state sync.
		// If this param is modified by governance, it will be reset by the upgrade module.
		types.NewParamSetPair(
			baseapp.ParamStoreKeyVersionParams, tmproto.VersionParams{}, baseapp.ValidateValidatorParams,
		),
	)
}
