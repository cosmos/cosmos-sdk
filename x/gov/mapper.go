package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"

	stake "github.com/cosmos/cosmos-sdk/x/dummy_stake"
)

type governanceMapper struct {
	// The reference to the CoinKeeper to modify balances
	ck bank.CoinKeeper

	// The reference to the StakeMapper to get information about stakers
	sm stake.Mapper

	// The (unexposed) keys used to access the stores from the Context.
	proposalStoreKey sdk.StoreKey

	// The wire codec for binary encoding/decoding.
	cdc *wire.Codec
}

// NewGovernanceMapper returns a mapper that uses go-wire to (binary) encode and decode gov types.
func NewGovernanceMapper(key sdk.StoreKey, ck bank.CoinKeeper) governanceMapper {
	cdc := wire.NewCodec()
	return governanceMapper{
		proposalStoreKey: key,
		ck:               ck,
		cdc:              cdc,
	}
}

// Returns the go-wire codec.
func (gm governanceMapper) WireCodec() *wire.Codec {
	return gm.cdc
}

func (gm governanceMapper) GetProposal(ctx sdk.Context, proposalID int64) *Proposal {
	store := ctx.KVStore(gm.proposalStoreKey)
	key, _ := gm.cdc.MarshalBinary(proposalID)
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	proposal := &Proposal{}
	err := gm.cdc.UnmarshalBinary(bz, proposal)
	if err != nil {
		panic(err)
	}

	return proposal
}

// Implements sdk.AccountMapper.
func (gm governanceMapper) SetProposal(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(gm.proposalStoreKey)

	bz, err := gm.cdc.MarshalBinary(proposal)
	if err != nil {
		panic(err)
	}

	key, _ := gm.cdc.MarshalBinary(proposal.ProposalID)

	store.Set(key, bz)
}

func (gm governanceMapper) getNewProposalID(ctx sdk.Context) int64 {
	store := ctx.KVStore(gm.proposalStoreKey)
	bz := store.Get([]byte("newProposalID"))
	if bz == nil {
		return 0
	}

	proposalID := new(int64)
	err := gm.cdc.UnmarshalBinary(bz, proposalID) // TODO: switch to UnmarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic("should not happen")
	}

	bz, err = gm.cdc.MarshalBinary(*proposalID + 1) // TODO: switch to MarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic("should not happen")
	}

	store.Set([]byte("newProposalID"), bz)

	return *proposalID
}

func (gm governanceMapper) getProposalQueue(ctx sdk.Context) ProposalQueue {
	store := ctx.KVStore(gm.proposalStoreKey)
	bz := store.Get([]byte("proposalQueue"))
	if bz == nil {
		return nil
	}

	proposalQueue := &ProposalQueue{}
	err := gm.cdc.UnmarshalBinary(bz, proposalQueue) // TODO: switch to UnmarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic(err)
	}

	return *proposalQueue
}

func (gm governanceMapper) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(gm.proposalStoreKey)

	bz, err := gm.cdc.MarshalBinary(proposalQueue) // TODO: switch to MarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic(err)
	}

	store.Set([]byte("proposalQueue"), bz)
}

func (gm governanceMapper) ProposalQueuePeek(ctx sdk.Context) *Proposal {
	proposalQueue := gm.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return gm.GetProposal(ctx, proposalQueue[0])
}

func (gm governanceMapper) ProposalQueuePop(ctx sdk.Context) *Proposal {
	proposalQueue := gm.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue := proposalQueue[0], proposalQueue[1:]
	gm.setProposalQueue(ctx, proposalQueue)
	return gm.GetProposal(ctx, frontElement)
}

func (gm governanceMapper) ProposalQueuePush(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(gm.proposalStoreKey)

	proposalQueue := append(gm.getProposalQueue(ctx), proposal.ProposalID)
	bz, err := gm.cdc.MarshalBinary(proposalQueue)
	if err != nil {
		panic(err)
	}
	store.Set([]byte("proposalQueue"), bz)
}

func (gm governanceMapper) GetActiveProcedure() *Procedure { // TODO: move to param store and allow for updating of this
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
