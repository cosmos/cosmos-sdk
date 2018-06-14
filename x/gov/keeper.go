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
	storeKey sdk.StoreKey

	// The wire codec for binary encoding/decoding.
	cdc *wire.Codec

	// Reserved codespace
	codespace sdk.CodespaceType
}

// NewGovernanceMapper returns a mapper that uses go-wire to (binary) encode and decode gov types.
func NewKeeper(cdc *wire.Codec, key sdk.StoreKey, ck bank.Keeper, sk stake.Keeper, codespace sdk.CodespaceType) Keeper {
	return Keeper{
		storeKey:  key,
		ck:        ck,
		sk:        sk,
		cdc:       cdc,
		codespace: codespace,
	}
}

// Returns the go-wire codec.
func (keeper Keeper) WireCodec() *wire.Codec {
	return keeper.cdc
}

// =====================================================
// Proposals

// Creates a NewProposal
func (keeper Keeper) NewProposal(ctx sdk.Context, title string, description string, proposalType string) *Proposal {
	proposal := &Proposal{
		ProposalID:       keeper.getNewProposalID(ctx),
		Title:            title,
		Description:      description,
		ProposalType:     proposalType,
		Status:           StatusDepositPeriod,
		TotalDeposit:     sdk.Coins{},
		SubmitBlock:      ctx.BlockHeight(),
		VotingStartBlock: -1, // TODO: Make Time
	}
	keeper.SetProposal(ctx, proposal)
	keeper.InactiveProposalQueuePush(ctx, proposal)
	return proposal
}

// Get Proposal from store by ProposalID
func (keeper Keeper) GetProposal(ctx sdk.Context, proposalID int64) *Proposal {
	store := ctx.KVStore(keeper.storeKey)
	key := []byte(fmt.Sprintf("%d", proposalID) + ":proposal")
	bz := store.Get(key)
	if bz == nil {
		return nil
	}

	proposal := &Proposal{}
	keeper.cdc.MustUnmarshalBinary(bz, proposal)

	return proposal
}

// Implements sdk.AccountMapper.
func (keeper Keeper) SetProposal(ctx sdk.Context, proposal *Proposal) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(proposal)
	key := []byte(fmt.Sprintf("%d", proposal.ProposalID) + ":proposal")
	store.Set(key, bz)
}

// Implements sdk.AccountMapper.
func (keeper Keeper) DeleteProposal(ctx sdk.Context, proposal *Proposal) {
	store := ctx.KVStore(keeper.storeKey)
	key := []byte(fmt.Sprintf("%d", proposal.ProposalID) + ":proposal")
	store.Delete(key)
}

func (keeper Keeper) getNewProposalID(ctx sdk.Context) (proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get([]byte("newProposalID"))
	if bz == nil {
		proposalID = 1
	} else {
		keeper.cdc.MustUnmarshalBinary(bz, &proposalID) // TODO: switch to UnmarshalBinaryBare when new go-amino gets added
	}
	bz = keeper.cdc.MustMarshalBinary(proposalID + 1) // TODO: switch to MarshalBinaryBare when new go-amino gets added
	store.Set([]byte("newProposalID"), bz)
	return proposalID
}

func (keeper Keeper) activateVotingPeriod(ctx sdk.Context, proposal *Proposal) {
	proposal.VotingStartBlock = ctx.BlockHeight()
	proposal.Status = StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)
	keeper.ActiveProposalQueuePush(ctx, proposal)
}

// =====================================================
// Procedures

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetDepositProcedure(ctx sdk.Context) *DepositProcedure {
	return &DepositProcedure{
		MinDeposit:       sdk.Coins{{"steak", 10}},
		MaxDepositPeriod: 200,
	}
}

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetVotingProcedure(ctx sdk.Context) *VotingProcedure {
	return &VotingProcedure{
		VotingPeriod: 200,
	}
}

// Gets procedure from store. TODO: move to global param store and allow for updating of this
func (keeper Keeper) GetTallyingProcedure(ctx sdk.Context) *TallyingProcedure {
	return &TallyingProcedure{
		Threshold:         sdk.NewRat(1, 2),
		Veto:              sdk.NewRat(1, 3),
		GovernancePenalty: sdk.NewRat(1, 100),
	}
}

// =====================================================
// Votes

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) AddVote(ctx sdk.Context, proposalID int64, voter sdk.Address, option string) sdk.Error {
	proposal := keeper.GetProposal(ctx, proposalID)
	if proposal == nil {
		return ErrUnknownProposal(proposalID)
	}
	if proposal.Status != StatusVotingPeriod {
		return ErrInactiveProposal(proposalID)
	}

	if option != "Yes" && option != "Abstain" && option != "No" && option != "NoWithVeto" {
		return ErrInvalidVote(option)
	}

	vote := Vote{
		ProposalID: proposalID,
		Voter:      voter,
		Option:     option,
	}
	keeper.setVote(ctx, proposalID, voter, vote)

	return nil
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) GetVote(ctx sdk.Context, proposalID int64, voter sdk.Address) *Vote {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get([]byte(fmt.Sprintf("%d", proposalID) + ":votes:" + fmt.Sprintf("%s", voter)))
	if bz == nil {
		return nil
	}
	vote := &Vote{}
	keeper.cdc.MustUnmarshalBinary(bz, vote)
	return vote
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) setVote(ctx sdk.Context, proposalID int64, voter sdk.Address, vote Vote) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(vote)
	key := []byte(fmt.Sprintf("%d", proposalID) + ":votes:" + fmt.Sprintf("%s", voter))
	store.Set(key, bz)
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) GetVotes(ctx sdk.Context, proposalID int64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(fmt.Sprintf("%d", proposalID)+":votes:"))
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) deleteVote(ctx sdk.Context, proposalID int64, voter sdk.Address) {
	store := ctx.KVStore(keeper.storeKey)
	key := []byte(fmt.Sprintf("%d", proposalID) + ":votes:" + fmt.Sprintf("%s", voter))
	store.Delete(key)
}

// =====================================================
// Deposits

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) GetDeposit(ctx sdk.Context, proposalID int64, depositer sdk.Address) *Deposit {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get([]byte(fmt.Sprintf("%d", proposalID) + ":deposits:" + fmt.Sprintf("%s", depositer)))
	if bz == nil {
		return nil
	}
	deposit := &Deposit{}
	keeper.cdc.MustUnmarshalBinary(bz, deposit)
	return deposit
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) setDeposit(ctx sdk.Context, proposalID int64, depositer sdk.Address, deposit Deposit) {
	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinary(deposit)
	key := []byte(fmt.Sprintf("%d", proposalID) + ":deposits:" + fmt.Sprintf("%s", depositer))
	store.Set(key, bz)
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) AddDeposit(ctx sdk.Context, proposalID int64, depositer sdk.Address, depositAmount sdk.Coins) sdk.Error {
	proposal := keeper.GetProposal(ctx, proposalID)
	if proposal == nil {
		return ErrUnknownProposal(proposalID)
	}

	if (proposal.Status != StatusDepositPeriod) && (proposal.Status != StatusVotingPeriod) {
		return ErrAlreadyFinishedProposal(proposalID)
	}

	_, _, err := keeper.ck.SubtractCoins(ctx, depositer, depositAmount)
	if err != nil {
		return err
	}

	proposal.TotalDeposit = proposal.TotalDeposit.Plus(depositAmount)
	keeper.SetProposal(ctx, proposal)
	if proposal.TotalDeposit.IsGTE(keeper.GetDepositProcedure(ctx).MinDeposit) {
		keeper.activateVotingPeriod(ctx, proposal)
	}

	currDeposit := keeper.GetDeposit(ctx, proposalID, depositer)
	if currDeposit == nil {
		newDeposit := Deposit{depositer, depositAmount}
		keeper.setDeposit(ctx, proposalID, depositer, newDeposit)
	} else {
		currDeposit.Amount = currDeposit.Amount.Plus(depositAmount)
		keeper.setDeposit(ctx, proposalID, depositer, *currDeposit)
	}

	return nil
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) GetDeposits(ctx sdk.Context, proposalID int64) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return sdk.KVStorePrefixIterator(store, []byte(fmt.Sprintf("%d", proposalID)+":deposits:"))
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) RefundDeposits(ctx sdk.Context, proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)

	for ; depositsIterator.Valid(); depositsIterator.Next() {
		deposit := &Deposit{}
		keeper.cdc.MustUnmarshalBinary(depositsIterator.Value(), deposit)

		_, _, err := keeper.ck.AddCoins(ctx, deposit.Depositer, deposit.Amount)
		if err != nil {
			panic("should not happen")
		}

		store.Delete(depositsIterator.Key())
	}

	depositsIterator.Close()
}

// Gets the vote of a specific voter on a specific proposal
func (keeper Keeper) DeleteDeposits(ctx sdk.Context, proposalID int64) {
	store := ctx.KVStore(keeper.storeKey)
	depositsIterator := keeper.GetDeposits(ctx, proposalID)

	for ; depositsIterator.Valid(); depositsIterator.Next() {
		store.Delete(depositsIterator.Key())
	}

	depositsIterator.Close()
}

// =====================================================
// ProposalQueues

func (keeper Keeper) getActiveProposalQueue(ctx sdk.Context) ProposalQueue {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get([]byte("activeProposalQueue"))
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

func (keeper Keeper) setActiveProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(keeper.storeKey)

	bz, err := keeper.cdc.MarshalBinary(proposalQueue) // TODO: switch to MarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic(err)
	}

	store.Set([]byte("activeProposalQueue"), bz)
}

// Return the Proposal at the front of the ProposalQueue
func (keeper Keeper) ActiveProposalQueuePeek(ctx sdk.Context) *Proposal {
	proposalQueue := keeper.getActiveProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return keeper.GetProposal(ctx, proposalQueue[0])
}

// Remove and return a Proposal from the front of the ProposalQueue
func (keeper Keeper) ActiveProposalQueuePop(ctx sdk.Context) *Proposal {
	proposalQueue := keeper.getActiveProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue := proposalQueue[0], proposalQueue[1:]
	keeper.setActiveProposalQueue(ctx, proposalQueue)
	return keeper.GetProposal(ctx, frontElement)
}

// Add a proposalID to the back of the ProposalQueue
func (keeper Keeper) ActiveProposalQueuePush(ctx sdk.Context, proposal *Proposal) {
	proposalQueue := append(keeper.getActiveProposalQueue(ctx), proposal.ProposalID)
	keeper.setActiveProposalQueue(ctx, proposalQueue)
}

func (keeper Keeper) getInactiveProposalQueue(ctx sdk.Context) ProposalQueue {
	store := ctx.KVStore(keeper.storeKey)
	bz := store.Get([]byte("inactiveProposalQueue"))
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

func (keeper Keeper) setInactiveProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(keeper.storeKey)

	bz, err := keeper.cdc.MarshalBinary(proposalQueue) // TODO: switch to MarshalBinaryBare when new go-amino gets added
	if err != nil {
		panic(err)
	}

	store.Set([]byte("inactiveProposalQueue"), bz)
}

// Return the Proposal at the front of the ProposalQueue
func (keeper Keeper) InactiveProposalQueuePeek(ctx sdk.Context) *Proposal {
	proposalQueue := keeper.getInactiveProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return keeper.GetProposal(ctx, proposalQueue[0])
}

// Remove and return a Proposal from the front of the ProposalQueue
func (keeper Keeper) InactiveProposalQueuePop(ctx sdk.Context) *Proposal {
	proposalQueue := keeper.getInactiveProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue := proposalQueue[0], proposalQueue[1:]
	keeper.setInactiveProposalQueue(ctx, proposalQueue)
	return keeper.GetProposal(ctx, frontElement)
}

// Add a proposalID to the back of the ProposalQueue
func (keeper Keeper) InactiveProposalQueuePush(ctx sdk.Context, proposal *Proposal) {
	proposalQueue := append(keeper.getInactiveProposalQueue(ctx), proposal.ProposalID)
	keeper.setInactiveProposalQueue(ctx, proposalQueue)
}
