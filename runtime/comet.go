package runtime

import (
	"context"

	"cosmossdk.io/core/abci"
	corecomet "cosmossdk.io/core/abci"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ corecomet.Service = &ContextAwareABCIInfoService{}

// ContextAwareABCIInfoService provides CometInfo which is embedded as a value in a Context.
// This the legacy (server v1, baseapp) way of accessing CometInfo at the module level.
type ContextAwareABCIInfoService struct{}

func (c ContextAwareABCIInfoService) ABCIInfo(ctx context.Context) abci.Info {
	ci := sdk.UnwrapSDKContext(ctx).CometInfo()
	return abci.Info{
		Evidence:        ci.Evidence,
		ValidatorsHash:  ci.ValidatorsHash,
		ProposerAddress: ci.ProposerAddress,
		LastCommit:      ci.LastCommit,
	}
}
