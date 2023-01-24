package sanction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// AccountKeeper defines the account/auth functionality needed from within the sanction module.
type AccountKeeper interface {
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
}

// BankKeeper defines the bank functionality needed from within the sanction module.
type BankKeeper interface {
	SetSanctionKeeper(keeper banktypes.SanctionKeeper)
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// GovKeeper defines the gov functionality needed from within the sanction module.
type GovKeeper interface {
	GetProposal(ctx sdk.Context, proposalID uint64) (govv1.Proposal, bool)
	GetDepositParams(ctx sdk.Context) govv1.DepositParams
	GetVotingParams(ctx sdk.Context) govv1.VotingParams
	GetProposalID(ctx sdk.Context) (uint64, error)
}
