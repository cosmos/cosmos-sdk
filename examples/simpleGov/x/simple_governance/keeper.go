package simpleGovernance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// nolint
type Keeper struct {
	SimpleGov sdk.StoreKey      // Key to our module's store
	codespace sdk.CodespaceType // Reserves space for error codes
	cdc       *wire.Codec       // Codec to encore/decode structs

	ck bank.Keeper  // Needed to handle deposits. This module onlyl requires read/writes to Atom balance
	sm stake.Keeper // Needed to compute voting power. This module only needs read access to the staking store
}

// NewKeeper crates a new keeper with write and read access
func NewKeeper(SimpleGov sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) Keeper {
	cdc := wire.NewCodec()

	return Keeper{
		SimpleGov: SimpleGov,
		cdc:       cdc,
		ck:        ck,
		sm:        sm,
		codespace: codespace,
	}
}

// NewProposalID creates a new id for a proposal
func (k Keeper) NewProposalID(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.SimpleGov)
	bid := store.Get([]byte("TotalID"))
	if bid == nil {
		return 0
	}

	totalID := new(int64)
	err := k.cdc.UnmarshalBinary(bid, totalID)
	if err != nil {
		panic(err)
	}

	return (*totalID + 1)
}

// GetProposal gets the proposal with the given id from the context.
func (k Keeper) GetProposal(ctx sdk.Context, proposalID int64) (Proposal, sdk.Error) {
	store := ctx.KVStore(k.SimpleGov)

	key := GenerateProposalKey(proposalID)
	bp := store.Get(key)
	if bp == nil {
		return Proposal{}, ErrProposalNotFound(proposalID)
	}

	proposal := Proposal{}

	err := k.cdc.UnmarshalBinary(bp, proposal)
	if err != nil {
		panic(err)
	}

	return proposal, nil
}

// GetAllProposals gets the set of all active proposals
// func (k Keeper) GetAllProposals(ctx sdk.Context) (proposals []Proposal) {
// 	storeKey := sdk.NewKVStoreKey("simpleGov")
// 	store := ctx.KVStore(storeKey)
// 	iterator := sdk.KVStorePrefixIterator(store, prefix) // Check the prefix key
//
// 	i := 0
// 	for ; ; i++ {
// 		if !iterator.Valid() {
// 			iterator.Close()
// 			break
// 		}
// 		bz := iterator.Value()
// 		var proposal Proposal
// 		k.cdc.MustUnmarshalBinary(bz, &proposal)
// 		if proposal.IsOpen() {
// 			proposals = append(proposals, proposal)
// 		}
// 		iterator.Next()
// 	}
// 	return proposals
// }

// SetProposal sets a proposal to the context
func (k Keeper) SetProposal(ctx sdk.Context, proposalID int64, proposal Proposal) sdk.Error {
	store := ctx.KVStore(k.SimpleGov)

	bp, err := k.cdc.MarshalBinary(proposal)
	if err != nil {
		panic(err) // return proper error
	}

	key := GenerateProposalKey(proposalID)

	store.Set(key, bp)
	return nil
}

// GetVote returns the given option of a proposal stored in the keeper
// Used to check if an address already voted
func (k Keeper) GetVote(ctx sdk.Context, proposalID int64, voter sdk.Address) (string, sdk.Error) {

	key := GenerateProposalVoteKey(proposalID, voter)
	store := ctx.KVStore(k.SimpleGov)
	bv := store.Get(key)
	if bv == nil {
		return "", ErrVoteNotFound()
	}
	option := new(string)
	err := k.cdc.UnmarshalBinary(bv, option)
	if err != nil {
		panic(err)
	}
	return *option, nil
}

// GetAllProposalVotes gets the set of all votes from a proposal
// func (k Keeper) GetAllProposalVotes(ctx sdk.Context, proposalID int64) (votes []VoteMsg) {
// 	storeKey := sdk.NewKVStoreKey("simpleGov")
// 	store := ctx.KVStore(storeKey)
// 	iterator := sdk.KVStorePrefixIterator(store, prefix) // Check the prefix key
//
// 	i := 0
// 	for ; ; i++ {
// 		if !iterator.Valid() {
// 			iterator.Close()
// 			break
// 		}
// 		bz := iterator.Value()
// 		var vote VoteMsg
// 		k.cdc.MustUnmarshalBinary(bz, &vote)
//
// 		votes = append(votes, vote)
// 		iterator.Next()
// 	}
// 	return votes
// }

// SetVote sets the vote option to the proposal stored in the context store
func (k Keeper) SetVote(ctx sdk.Context, proposalID int64, voter sdk.Address, option string) {
	key := GenerateProposalVoteKey(proposalID, voter)
	store := ctx.KVStore(k.SimpleGov)
	bv, err := k.cdc.MarshalBinary(option)
	if err != nil {
		panic(err)
	}
	store.Set(key, bv)
}

//--------------------------------------------------------------------------------------

// getProposalQueue gets the ProposalQueue from the context
func (k Keeper) getProposalQueue(ctx sdk.Context) (ProposalQueue, sdk.Error) {
	store := ctx.KVStore(k.SimpleGov)
	bpq := store.Get([]byte("proposalQueue"))
	if bpq == nil {
		return ProposalQueue{}, ErrProposalQueueNotFound()
	}

	proposalQueue := ProposalQueue{}
	err := k.cdc.UnmarshalBinaryBare(bpq, proposalQueue)
	if err != nil {
		panic(err)
	}

	return proposalQueue, nil
}

// setProposalQueue sets the ProposalQueue to the context
func (k Keeper) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(k.SimpleGov)
	bpq, err := k.cdc.MarshalBinaryBare(proposalQueue)
	if err != nil {
		panic(err)
	}
	store.Set([]byte("proposalQueue"), bpq)
}

// ProposalQueueHead returns the head of the FIFO Proposal queue
func (k Keeper) ProposalQueueHead(ctx sdk.Context) (Proposal, sdk.Error) {
	proposalQueue, err := k.getProposalQueue(ctx)
	if err != nil {
		return Proposal{}, err
	}
	if proposalQueue.isEmpty() {
		return Proposal{}, ErrEmptyProposalQueue()
	}
	proposal, err := k.GetProposal(ctx, proposalQueue[0])
	if err != nil {
		return Proposal{}, err
	}
	return proposal, nil
}

// ProposalQueuePop pops the head from the Proposal queue
func (k Keeper) ProposalQueuePop(ctx sdk.Context) (Proposal, sdk.Error) {
	proposalQueue, err := k.getProposalQueue(ctx)
	if err != nil {
		return Proposal{}, err
	}
	if proposalQueue.isEmpty() {
		return Proposal{}, ErrEmptyProposalQueue()
	}
	headElement, tailProposalQueue := proposalQueue[0], proposalQueue[1:]
	k.setProposalQueue(ctx, tailProposalQueue)
	proposal, err := k.GetProposal(ctx, headElement)
	if err != nil {
		return Proposal{}, err
	}
	return proposal, nil
}

// ProposalQueuePush pushes a proposal to the tail of the FIFO Proposal queue
func (k Keeper) ProposalQueuePush(ctx sdk.Context, proposaID int64) sdk.Error {
	proposalQueue, err := k.getProposalQueue(ctx)
	if err != nil {
		return err
	}
	proposalQueue = append(proposalQueue, proposaID)
	k.setProposalQueue(ctx, proposalQueue)
	return nil
}

//--------------------------------------------------------------------------------------

// KeeperRead is a Keeper only with read access
type KeeperRead struct {
	Keeper
}

// NewKeeperRead crates a new keeper with read access
func NewKeeperRead(SimpleGov sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) KeeperRead {
	cdc := wire.NewCodec()

	return KeeperRead{Keeper{
		SimpleGov: SimpleGov,
		cdc:       cdc,
		ck:        ck,
		sm:        sm,
		codespace: codespace,
	}}
}

// NewProposalID creates a new id for a proposal
func (k KeeperRead) NewProposalID(ctx sdk.Context) sdk.Error {
	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
}

// SetProposal sets a proposal to the context
func (k KeeperRead) SetProposal(ctx sdk.Context, proposalID int64, proposal Proposal) sdk.Error {
	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
}

// SetVote sets the vote option to the proposal stored in the context store
func (k KeeperRead) SetVote(ctx sdk.Context, key []byte, option string) sdk.Error {
	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
}

// setProposalQueue sets the ProposalQueue to the context
func (k KeeperRead) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) sdk.Error {
	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
}
