package simple_governance

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos/sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/wire"
	stake "github.com/cosmos/cosmos-sdk/examples/simpleGov/x/simplestake"
)

type Keeper struct {
	ProposalStoreKey 	sdk.StoreKey
	codespace 			sdk.CodespaceType
	Cdc              	*wire.Codec

	ck               	bank.Keeper
	sm               	stake.Keeper
}

type KeeperRead struct {
	Keeper
}


func NewKeeper(proposalStoreKey sdk.StoreKey, ck bank.CoinKeeper, sm stake.KeeperRead, codespace sdk.CodespaceType) Keeper {
	cdc = wire.NewCodec()
	
	return Keeper{
		ProposalStoreKey: 	proposalStoreKey,
		Cdc: 				cdc,
		ck: 				ck,
		sm: 				sm,
		codespace: 			codespace,
	}
}

func NewKeeperRead(proposalStoreKey sdk.StoreKey, ck bank.CoinKeeper, sm stake.KeeperRead, codespace sdk.CodespaceType) KeeperRead {
	cdc = wire.NewCodec()

	return KeeperRead{
		ProposalStoreKey: 	proposalStoreKey,
		Cdc: 				cdc,
		ck: 				ck,
		sm: 				sm,
		codespace: 			codespace,
	}
}

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

func (k KeeperRead) SetProposal(ctx sdk.Context, proposalID int64, proposal Proposal) sdk.Error {
	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
}

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

func (k Keeper) SetOption(ctx sdk.Context, key []byte, option string) sdk.Error {
	store := ctx.KVStore(k.proposalStoreKey)

	bv, err := k.cdc.MarshalBinary(option)
	if err != nil {
		panic(error)
	}

	store.set(key, bv)
	return nil
}

func (k KeeperRead) SetOption(ctx sdk.Context, key []byte, option string) sdk.Error {
	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
}

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
}

func (k KeeperRead) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) sdk.Error {
	return sdk.ErrUnauthorized("").Trace("This keeper does not have write access for the simple governance store")
}


func (k Keeper) ProposalQueuePeek(ctx sdk.Context) Proposal {
	proposalQueue := k.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return k.GetProposal(ctx, proposalQueue[0])
}

func (k Keeper) ProposalQueuePop(ctx sdk.Context) Proposal {
	proposalQueue := k.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue = proposalQueue[0], proposalQueue[1:]
	k.setProposalQueue(ctx, proposalQueue)
	return k.GetProposal(ctx, frontElement)
}

func (k Keeper) ProposalQueuePush(ctx sdk.Context, proposaID int64) {
	proposalQueue := append(k.getProposalQueue(ctx), proposalID)
	k.setProposalQueue(ctx, proposalQueue)
}
