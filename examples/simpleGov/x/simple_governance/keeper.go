package simpleGovernance

import (
	stake "github.com/cosmos/cosmos-sdk/examples/simpleGov/x/simplestake"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	bank "github.com/cosmos/cosmos/sdk/x/bank"
)

// nolint
type Keeper struct {
	ProposalStoreKey sdk.StoreKey
	codespace        sdk.CodespaceType
	cdc              *wire.Codec

	ck bank.Keeper
	sm stake.Keeper
}

// nolint
type KeeperRead struct {
	Keeper
}

// NewKeeper crates a new keeper with write and read access
func NewKeeper(proposalStoreKey sdk.StoreKey, ck bank.CoinKeeper, sm stake.KeeperRead, codespace sdk.CodespaceType) Keeper {
	cdc = wire.NewCodec()

	return Keeper{
		ProposalStoreKey: proposalStoreKey,
		Cdc:              cdc,
		ck:               ck,
		sm:               sm,
		codespace:        codespace,
	}
}

// NewKeeperRead crates a new keeper with read access
func NewKeeperRead(proposalStoreKey sdk.StoreKey, ck bank.CoinKeeper, sm stake.KeeperRead, codespace sdk.CodespaceType) KeeperRead {
	cdc = wire.NewCodec()

	return KeeperRead{
		ProposalStoreKey: proposalStoreKey,
		Cdc:              cdc,
		ck:               ck,
		sm:               sm,
		codespace:        codespace,
	}
}

// GetProposal gets the proposal with the given id from the context
func (k Keeper) GetProposal(ctx sdk.Context, proposalID int64) Proposal {
	store := ctx.KVStore(k.ProposalStoreKey)

	bpi, err := k.cdc.MarshalBinary(proposalID)
	if err != nil {
		panic(error)
	}

	bp = store.Get(bpi)
	if bp == nil {
		return nil
	}

	proposal := Proposal{}

	err := k.cdc.UnmarshalBinary(bp, proposal)
	if err != nil {
		panic(error)
	}

	return proposal
}

// SetProposal sets a proposal to the context
func (k Keeper) SetProposal(ctx sdk.Context, proposalID int64, proposal Proposal) sdk.Error {
	store := ctx.KVStore(k.ProposalStoreKey)

	bp, err := k.cdc.MarshalBinary(proposal)
	if err != nil {
		panic(error) // return proper error
	}

	bpi, err := k.cdc.MarshalBinary(proposalID)
	if err != nil {
		panic(error) // return proper error
	}

	store.set(bpi, bp)
	return nil
}

// func (k KeeperRead) SetProposal(ctx sdk.Context, proposalID int64, proposal Proposal) sdk.Error {
// 	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
// }

// NewProposalID creates a new id for a proposal
func (k Keeper) NewProposalID(ctx sdk.Context) int64 {
	store := ctx.KVStore(k.ProposalStoreKey)

	bid := store.Get([]byte("TotalID"))
	if bid == nil {
		return 0
	}

	totalID := new(int64)
	err := k.cdc.UnmarshalBinary(bid, totalID)
	if err != nil {
		panic(error)
	}

	return (totalID + 1)
}

//--------------------------------------------------------------------------------------

// GetOption returns the given option of a proposal stored in the keeper
func (k Keeper) GetOption(ctx sdk.Context, key []byte) string {
	store := ctx.KVStore(k.proposalStoreKey)

	bv = store.Get(key)
	if bv == nil {
		return nil
	}

	option := new(string)

	err := k.cdc.UnmarshalBinary(bv, option)
	if err != nil {
		panic(error)
	}

	return option
}

// SetOption sets the option to the propposal stored in the context store
func (k Keeper) SetOption(ctx sdk.Context, key []byte, option string) {
	store := ctx.KVStore(k.proposalStoreKey)

	bv, err := k.cdc.MarshalBinary(option)
	if err != nil {
		panic(error)
	}

	store.set(key, bv)
	return nil
}

// IMO not even necessary
// func (k KeeperRead) SetOption(ctx sdk.Context, key []byte, option string) sdk.Error {
// 	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
// }

//--------------------------------------------------------------------------------------

func (k Keeper) getProposalQueue(ctx sdk.Context) ProposalQueue {
	store := ctx.KVStore(k.proposalStoreKey)
	bpq := store.Get([]byte("proposalQueue"))
	if bz == nil {
		return nil
	}

	proposalQueue := &ProposalQueue{}
	err := k.cdc.UnmarshalBinaryBare(bpq, proposalQueue)
	if err != nil {
		panic(err)
	}

	return proposalQueue
}

func (k Keeper) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(k.proposalStoreKey)
	bpq, err := k.cdc.MarshalBinaryBare(proposalQueue)
	if err != nil {
		panic(err)
	}
	store.Set([]byte("proposalQueue"), bpq)
	return nil
}

// func (k KeeperRead) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) sdk.Error {
// 	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
// }

// ProposalQueueHead returns the head of the FIFO Proposal queue
func (k Keeper) ProposalQueueHead(ctx sdk.Context) (Proposal, sdk.Error) {
	proposalQueue := k.getProposalQueue(ctx)
	if proposalQueue.IsEmpty() {
		return nil, ErrEmptyProposalQueue()
	}
	return k.GetProposal(ctx, proposalQueue[0]), nil
}

// ProposalQueuePop pops the head from the Proposal queue
func (k Keeper) ProposalQueuePop(ctx sdk.Context) (Proposal, sdk.Error) {
	proposalQueue := k.getProposalQueue(ctx)
	if proposalQueue.IsEmpty() {
		return nil, ErrEmptyProposalQueue()
	}
	headElement, tailProposalQueue = proposalQueue[0], proposalQueue[1:]
	err := k.setProposalQueue(ctx, tailProposalQueue)
	if err != nil {
		return nil, err
	}
	return k.GetProposal(ctx, headElement), nil
}

// ProposalQueuePush pushes a proposal to the tail of the FIFO Proposal queue
func (k Keeper) ProposalQueuePush(ctx sdk.Context, proposaID int64) {
	proposalQueue := append(k.getProposalQueue(ctx), proposalID)
	k.setProposalQueue(ctx, proposalQueue)
}
