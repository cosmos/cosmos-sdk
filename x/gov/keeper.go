package gov

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"

	stake "github.com/cosmos/cosmos-sdk/x/stake"
)

// Governance Keeper
type Keeper struct {
	// The reference to the CoinKeeper to modify balances
	ck bank.Keeper

	// The reference to the StakeMapper to get information about stakers
	sk stake.Keeper

	// The (unexposed) keys used to access the stores from the Context.
	proposalStoreKey sdk.StoreKey

	// The wire codec for binary encoding/decoding.
	cdc *wire.Codec

	// Reserved codespace
	codespace sdk.CodespaceType
}

// NewGovernanceMapper returns a mapper that uses go-wire to (binary) encode and decode gov types.
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, sk stake.Keeper, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		proposalStoreKey: key,
		ck:               ck,
		sk:               sk,
		cdc:              cdc,
		codespace:        codespace,
	}
}

// Returns the go-wire codec.
func (keeper Keeper) WireCodec() *wire.Codec {
	return keeper.cdc
}

// Get Proposal from store by ProposalID
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
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal *Proposal) {
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

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetDepositProcedure() *DepositProcedure {
	return &DepositProcedure{
		MinDeposit:       sdk.Coins{{"atom", 10}},
		MaxDepositPeriod: 200,
	}
}

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetVotingProcedure() *VotingProcedure {
	return &VotingProcedure{
		VotingPeriod: 200,
	}
}

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetTallyingProcedure() *TallyingProcedure {
	return &TallyingProcedure{
		Threshold:         sdk.NewRat(1, 2),
		Veto:              sdk.NewRat(1, 3),
		GovernancePenalty: sdk.NewRat(1, 100),
	}
}

func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, proposal *Proposal) {
	proposal.VotingStartBlock = ctx.BlockHeight()
	keeper.SetProposal(ctx, proposal)
	keeper.ProposalQueuePush(ctx, proposal)
}

// Creates a NewProposal
func (keeper Keeper) NewProposal(ctx sdk.Context, title string, description string, proposalType string, initDeposit Deposit) (Proposal, sdk.Error) { // TODO: move to param store and allow for updating of this
	return Proposal{}, nil
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID int64, voter sdk.Address) *Vote {
	store := ctx.KVStore(keeper.proposalStoreKey)
	bz := store.Get([]byte(fmt.Sprintf("%d", proposalID) + ":" + fmt.Sprintf("%s", voter)))
	if bz == nil {
		return nil
	}
	vote := &Vote{}
	keeper.cdc.MustUnmarshalBinary(bz, vote)
	return vote
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) setVote(ctx sdk.Context, proposalID int64, voter sdk.Address, vote Vote) {
	store := ctx.KVStore(keeper.proposalStoreKey)
	bz := keeper.cdc.MustMarshalBinary(vote)
	key := []byte(fmt.Sprintf("%d", proposalID) + ":" + fmt.Sprintf("%s", voter))
	store.Set(key, bz)
}

// =====================================================
// ProposalQueue

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

// Return the Proposal at the front of the ProposalQueue
func (keeper Keeper) ProposalQueuePeek(ctx sdk.Context) *Proposal {
	proposalQueue := keeper.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return keeper.GetProposal(ctx, proposalQueue[0])
}

// Remove and return a Proposal from the front of the ProposalQueue
func (keeper Keeper) ProposalQueuePop(ctx sdk.Context) *Proposal {
	proposalQueue := keeper.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue := proposalQueue[0], proposalQueue[1:]
	keeper.setProposalQueue(ctx, proposalQueue)
	return keeper.GetProposal(ctx, frontElement)
}

// Add a proposalID to the back of the ProposalQueue
func (keeper Keeper) ProposalQueuePush(ctx sdk.Context, proposal *Proposal) {
	store := ctx.KVStore(keeper.proposalStoreKey)
	proposalQueue := append(keeper.getProposalQueue(ctx), proposal.ProposalID)
	bz := keeper.cdc.MustMarshalBinary(proposalQueue)
	store.Set([]byte("proposalQueue"), bz)
}
