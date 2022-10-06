package keeper

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

// HandleUpdateDenomMetadataProposal handles update of msg fee denom metadata
func HandleUpdateDenomMetadataProposal(ctx sdk.Context, k Keeper, proposal *types.UpdateDenomMetadataProposal, registry codectypes.InterfaceRegistry) error {
	if err := sdk.ValidateDenom(proposal.Metadata.Base); err != nil {
		return err
	}

	if err := proposal.Metadata.Validate(); err != nil {
		return err
	}

	k.SetDenomMetaData(ctx, proposal.Metadata)
	return nil
}
