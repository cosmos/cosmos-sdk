package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"

	stake "github.com/cosmos/cosmos-sdk/x/stake"
)

type Keeper struct {
	// The reference to the CoinKeeper to modify balances
	ck bank.CoinKeeper

	// The reference to the StakeMapper to get information about stakers
	sm stake.Keeper

	// The (unexposed) keys used to access the stores from the Context.
	proposalStoreKey sdk.StoreKey

	// The wire codec for binary encoding/decoding.
	cdc *wire.Codec
}

// NewGovernanceMapper returns a mapper that uses go-wire to (binary) encode and decode gov types.
func NewKeeper(key sdk.StoreKey, ck bank.CoinKeeper, sk stake.Keeper) Keeper {
	cdc := wire.NewCodec()
	return Keeper{
		proposalStoreKey: key,
		ck:               ck,
		cdc:              cdc,
	}
}

// Returns the go-wire codec.
func (keeper Keeper) WireCodec() *wire.Codec {
	return keeper.cdc
}

func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID int64) *Proposal {
	store := ctx.KVStore(keeper.proposalStoreKey)
	key, _ := keeper.cdc.MarshalBinary(proposalID)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	proposal := &Proposal{}
	err := keeper.cdc.UnmarshalBinary(bz, proposal)
	if err != nil {
		panic(err)
	}

	return proposal
}

// Implements sdk.AccountMapper.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(keeper.proposalStoreKey)

	bz, err := keeper.cdc.MarshalBinary(proposal)
	if err != nil {
		panic(err)
	}

	key, _ := keeper.cdc.MarshalBinary(proposal.ProposalID)

	store.Set(key, bz)
}

func (keeper Keeper) getNewProposalID(ctx sdk.Context) int64 {
	store := ctx.KVStore(keeper.proposalStoreKey)
	bz := store.Get([]byte("newProposalID"))
	if bz == nil {
		return 0
	}

	proposalID := new(int64)
	err := keeper.cdc.UnmarshalBinary(bz, proposalID) // TODO: switch to UnmarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic("should not happen")
	}

	bz, err = keeper.cdc.MarshalBinary(*proposalID + 1) // TODO: switch to MarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic("should not happen")
	}

	store.Set([]byte("newProposalID"), bz)

	return *proposalID
}

func (keeper Keeper) getProposalQueue(ctx sdk.Context) ProposalQueue {
	store := ctx.KVStore(keeper.proposalStoreKey)
	bz := store.Get([]byte("proposalQueue"))
	if bz == nil {
		return nil
	}

	proposalQueue := &ProposalQueue{}
	err := keeper.cdc.UnmarshalBinary(bz, proposalQueue) // TODO: switch to UnmarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic(err)
	}

	return *proposalQueue
}

func (keeper Keeper) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(keeper.proposalStoreKey)

	bz, err := keeper.cdc.MarshalBinary(proposalQueue) // TODO: switch to MarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic(err)
	}

	store.Set([]byte("proposalQueue"), bz)
}

func (keeper Keeper) ProposalQueuePeek(ctx sdk.Context) *Proposal {
	proposalQueue := keeper.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return keeper.GetProposal(ctx, proposalQueue[0])
}

func (keeper Keeper) ProposalQueuePop(ctx sdk.Context) *Proposal {
	proposalQueue := keeper.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue := proposalQueue[0], proposalQueue[1:]
	keeper.setProposalQueue(ctx, proposalQueue)
	return keeper.GetProposal(ctx, frontElement)
}

func (keeper Keeper) ProposalQueuePush(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(keeper.proposalStoreKey)

	proposalQueue := append(keeper.getProposalQueue(ctx), proposal.ProposalID)
	bz, err := keeper.cdc.MarshalBinary(proposalQueue)
	if err != nil {
		panic(err)
	}
	store.Set([]byte("proposalQueue"), bz)
}

func (keeper Keeper) GetActiveProcedure() *Procedure { // TODO: move to param store and allow for updating of this
	return &Procedure{
		VotingPeriod:      200,
		MinDeposit:        sdk.Coins{{"atom", 10}},
		ProposalTypes:     []string{"TextProposal"},
		Threshold:         sdk.NewRat(1, 2),
		Veto:              sdk.NewRat(1, 3),
		MaxDepositPeriod:  200,
		GovernancePenalty: sdk.NewRat(1, 100),
	}
}

func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, proposal *Proposal) {
	proposal.VotingStartBlock = ctx.BlockHeight()

	// TODO: Can we get this directly from stakeState
	// stakeState := k.sm.loadGlobalState()
	// proposal.TotalVotingPower = stakeState.TotalSupply

	proposal.TotalVotingPower = 0

	validatorList := keeper.sm.GetValidators(ctx, 100) // TODO: Finalize with staking module

	for index, validator := range validatorList {
		validatorGovInfo := ValidatorGovInfo{
			ProposalID:      proposal.ProposalID,
			ValidatorAddr:   validator.Address,     // TODO: Finalize with staking module
			InitVotingPower: validator.VotingPower, // TODO: Finalize with staking module
			Minus:           0,
			LastVoteWeight:  -1,
		}

		proposal.ValidatorGovInfos = append(proposal.ValidatorGovInfos, validatorGovInfo)
	}

	keeper.ProposalQueuePush(ctx, *proposal)
}

func (keeper Keeper) NewProposal(ctx, title string, description string, proposalType string, initDeposit Deposit) (Proposal, sdk.Error) { // TODO: move to param store and allow for updating of this

}
