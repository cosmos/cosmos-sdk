package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

// ClientUpdateProposal will update the client if the propoal passes.
func (k Keeper) ClientUpdateProposal(ctx sdk.Context, p *ibctypes.ClientUpdateProposal) error {
	clientState, found := k.GetClientState(ctx, p.ClientId)
	if !found {
		return types.ErrClientNotFound
	}

	clientType := clientState.ClientType()
	switch clientType {
	case exported.Tendermint:
		{
		}
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
			tmtHeader, err := ibctmtypes.UnmarshalHeader(p.Header)
			if err != nil {
				return types.ErrInvalidHeader
			}
			if _, err = k.UpdateClient(ctx, p.ClientId, tmtHeader, true); err != nil {
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
