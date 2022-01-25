package slashing

import (
	"github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// NewSlashProposalsHandler  creates a new governance Handler for a slashing proposals
func NewSlashProposalsHandler(slashingKeeper keeper.Keeper, stakingKeeper stakingkeeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.SlashValidatorProposal:
			return handleSlashValidatorProposal(ctx, c, slashingKeeper, stakingKeeper)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized param proposal content type: %T", c)
		}
	}
}

// todo write test
func handleSlashValidatorProposal(
	ctx sdk.Context, p *types.SlashValidatorProposal,
	slashingKeeper keeper.Keeper, stakingKeeper stakingkeeper.Keeper,
) error {
	// todo validate if validator exists and active or unbonding(has unbound time before vote ends?).
	// maybe it is possible to validate somewhere before it passed?

	valAddr, valAddrErr := sdk.ValAddressFromBech32(p.ValidatorAddress)
	if valAddrErr != nil {
		return valAddrErr
	}
	validator, found := stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return stakingtypes.ErrNoValidatorFound
	}
	valConsAddr, valConsdAddrErr := validator.GetConsAddr()
	if valConsdAddrErr != nil {
		return valConsdAddrErr
	}

	// todo what if unbouding time is less than vote period?
	// todo slash using current power and delegations and redelegations?
	// or maybe allowing delegators to re-delegate during vote period?
	// lets consider all approaches
	valPower := stakingKeeper.GetLastValidatorPower(ctx, valAddr)
	slashingKeeper.Slash(ctx, valConsAddr, p.SlashFactor, valPower, ctx.BlockHeight())
	return nil
}
