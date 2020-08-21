package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// ClientUpdateProposal will try to update the client with the new header if and only if
// the propoal passes and one of the following two conditions is fulfill,
// 		1) AllowGovernanceOverrideAfterExpiry=true and Expire(ctx.BlockTime) = true
// 		2) AllowGovernanceOverrideAfterMinbehaviour and IsFrozen() = true
// In case 2) before trying to update the client, the client will be unfreeze by calling Unfreeze().
// Note, that even if the update happens, there is no garantie to ensure that it will be successful.
// The reason is that there are still some validation checks on the header that may fail and
// throw and error before the update is completed.
func (k Keeper) ClientUpdateProposal(ctx sdk.Context, p *types.ClientUpdateProposal) error {
	clientState, found := k.GetClientState(ctx, p.ClientId)
	if !found {
		return types.ErrClientNotFound
	}

	clientType := clientState.ClientType()
	switch clientType {
	case exported.Tendermint:

		tmClientState := clientState.(*ibctmtypes.ClientState)

		updateClientFlag := false
		if tmClientState.AllowGovernanceOverrideAfterExpiry && tmClientState.Expired(ctx.BlockTime()) {
			updateClientFlag = true
		}

		if tmClientState.AllowGovernanceOverrideAfterMisbehaviour && tmClientState.IsFrozen() {
			tmClientState.Unfreeze()
			k.SetClientState(ctx, p.ClientId, tmClientState)
			updateClientFlag = true
		}

		if updateClientFlag {

			var tmtHeader ibctmtypes.Header
			h, err := tmtHeader.UnmarshalBinaryBare(p.Header)
			if err != nil {
				return types.ErrInvalidHeader
			}
			if _, err = k.UpdateClient(ctx, p.ClientId, h, true); err != nil {
				return err
			}

		} else {
			return types.ErrFailUpdateClient
		}

	default:
		return sdkerrors.Wrapf(types.ErrInvalidClientType, "unsupported client type (%s)", clientType)
	}

	return nil
}
