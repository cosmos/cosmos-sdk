package simpleGovernance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/stake"
	amino "github.com/tendermint/go-amino"
)

// nolint
type Keeper struct {
	storeKey  sdk.StoreKey      // Key to our module's store
	codespace sdk.CodespaceType // Reserves space for error codes
	cdc       *wire.Codec       // Codec to encore/decode structs

	ck bank.Keeper  // Needed to handle deposits. This module onlyl requires read/writes to Atom balance
	sm stake.Keeper // Needed to compute voting power. This module only needs read access to the staking store
}

// NewKeeper crates a new keeper with write and read access
func NewKeeper(cdc *amino.Codec, simpleGovKey sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) Keeper {

	return Keeper{
		storeKey:  simpleGovKey,
		cdc:       cdc,
		ck:        ck,
		sm:        sm,
		codespace: codespace,
	}
}

// Creates a new Proposal
func (k Keeper) NewProposal(ctx sdk.Context, title string, description string) Proposal {
	proposalID := k.newProposalID(ctx)
	sdk.AccAddressFromHex("0")
	proposal := Proposal{
		ID:          proposalID,
		Title:       title,
		Description: description,
		State:       "Open",
		SubmitBlock: ctx.BlockHeight(),
	}
	k.SetProposal(ctx, proposal)
	k.ProposalQueuePush(ctx, proposal.ID)
	return proposal
}

// generates a new id for a proposal
func (k Keeper) newProposalID(ctx sdk.Context) (proposalID int64) {
	store := ctx.KVStore(k.storeKey)
	bid := store.Get(KeyNextProposalID)
	if bid == nil {
		return -1
	}
	k.cdc.MustUnmarshalBinary(bid, &proposalID)
	bid = k.cdc.MustMarshalBinary(proposalID + 1)
	store.Set(KeyNextProposalID, bid)
	return
}

// GetProposal gets the proposal with the given id from the context.
func (k Keeper) GetProposal(ctx sdk.Context, proposalID int64) (Proposal, sdk.Error) {
	store := ctx.KVStore(k.storeKey)

	key := GenerateProposalKey(proposalID)
	bp := store.Get(key)
	if bp == nil {
		return Proposal{}, ErrProposalNotFound(proposalID)
	}
	proposal := &Proposal{}
	k.cdc.MustUnmarshalBinary(bp, proposal)

	return *proposal, nil
}

// SetProposal sets a proposal to the context
func (k Keeper) SetProposal(ctx sdk.Context, proposal Proposal) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinary(proposal)
	key := GenerateProposalKey(proposal.ID)
	store.Set(key, bz)
}

// GetVote returns the given option of a proposal stored in the keeper
// Used to check if an address already voted
func (k Keeper) GetVote(ctx sdk.Context, proposalID int64, voter sdk.AccAddress) (option string, err sdk.Error) {

	key := GenerateProposalVoteKey(proposalID, voter)
	store := ctx.KVStore(k.storeKey)
	bv := store.Get(key)
	if bv == nil {
		return "", ErrVoteNotFound("")
	}
	k.cdc.MustUnmarshalBinary(bv, &option)
	return option, nil
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
func (k Keeper) SetVote(ctx sdk.Context, proposalID int64, voterAddr sdk.AccAddress, option string) {
	key := GenerateProposalVoteKey(proposalID, voterAddr)
	store := ctx.KVStore(k.storeKey)
	bv, err := k.cdc.MarshalBinary(option)
	if err != nil {
		panic(err)
	}
	store.Set(key, bv)
}

//--------------------------------------------------------------------------------------

// getProposalQueue gets the ProposalQueue from the context
func (k Keeper) getProposalQueue(ctx sdk.Context) (ProposalQueue, sdk.Error) {
	store := ctx.KVStore(k.storeKey)
	bpq := store.Get([]byte("proposalQueue"))
	if bpq == nil {
		return ProposalQueue{}, ErrProposalQueueNotFound("")
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
	store := ctx.KVStore(k.storeKey)
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
		return Proposal{}, ErrEmptyProposalQueue("Can't get element from an empty proposal queue")
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
		return Proposal{}, ErrEmptyProposalQueue("Can't get element from an empty proposal queue")
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
func NewKeeperRead(cdc *amino.Codec, simpleGovKey sdk.StoreKey, ck bank.Keeper, sm stake.Keeper, codespace sdk.CodespaceType) KeeperRead {
	return KeeperRead{Keeper{
		storeKey:  simpleGovKey,
		cdc:       cdc,
		ck:        ck,
		sm:        sm,
		codespace: codespace,
	}}
}

// NewProposalID creates a new id for a proposal
func (k KeeperRead) NewProposalID(ctx sdk.Context) sdk.Error {
	return sdk.ErrUnauthorized("").TraceSDK("This keeper does not have write access for the simple governance store")
}

// SetProposal sets a proposal to the context
func (k KeeperRead) SetProposal(ctx sdk.Context, proposal Proposal) sdk.Error {
	return sdk.ErrUnauthorized("").TraceSDK("This keeper does not have write access for the simple governance store")
}

// SetVote sets the vote option to the proposal stored in the context store
func (k KeeperRead) SetVote(ctx sdk.Context, key []byte, option string) sdk.Error {
	return sdk.ErrUnauthorized("").TraceSDK("This keeper does not have write access for the simple governance store")
}

// setProposalQueue sets the ProposalQueue to the context
func (k KeeperRead) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) sdk.Error {
	return sdk.ErrUnauthorized("").TraceSDK("This keeper does not have write access for the simple governance store")
}
