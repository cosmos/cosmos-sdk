package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// ClientUpdateProposal will try to update the client with the new header if and only if
// the propoal passes and one of the following two conditions is satisfied:
// 		1) AllowGovernanceOverrideAfterExpiry=true and Expire(ctx.BlockTime) = true
// 		2) AllowGovernanceOverrideAfterMinbehaviour and IsFrozen() = true
// In case 2) before trying to update the client, the client will be unfreeze by calling Unfreeze().
// Note, that even if the update happens, it may not be successful.
// The reason is that there are still some validation checks on the header that may fail and
// throw and error before the update is completed.
func (k Keeper) ClientUpdateProposal(ctx sdk.Context, p *types.ClientUpdateProposal) error {
	clientState, found := k.GetClientState(ctx, p.ClientId)
	if !found {
		return types.ErrClientNotFound
	}

	// TODO: it will be nice to refactor the following piece of code
	// and add a function to the clientStatus interface in order to
	// avoid to use the switch statement. A naive idea is to add
	// a function with following signature:
	//
	// TryClientUpdateProposal(sdk.Context, k keeper.Keeper, *types.ClientUpdateProposal) error
	//
	// Unfortunately, we can't pursuit this route as of now, because we face the challenge of an import cycle
	// with the keeper package which is not easy to resolve.
	clientType := clientState.ClientType()
	switch clientType {
	case exported.Tendermint:

		tmClientState := clientState.(*ibctmtypes.ClientState)

		updateClientFlag := false
		overrideFlag := false

		if tmClientState.AllowGovernanceOverrideAfterExpiry && tmClientState.Expired(ctx.BlockTime()) {
			updateClientFlag = true
			overrideFlag = true
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
			if _, err = k.UpdateClient(ctx, p.ClientId, h, overrideFlag); err != nil {
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
