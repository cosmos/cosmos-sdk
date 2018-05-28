-package module_tutorial

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos/sdk/x/bank"
	amino "github.com/tendermint/go-amino"
	stake "github.coms/cosmos/cosmos-sdk/x/stake"
)

type SimpleGovernanceKeeper struct {
	ProposalStoreKey sdk.StoreKey
	Cdc              *amino.Codec
	WriteAccess      bool
	ck               bank.CoinKeeper
	sm               stake.KeeperRead
}

func NewSimpleGovernanceKeeper(proposalStoreKey sdk.StoreKey, ck bank.CoinKeeper, sm stake.KeeperRead) SimpleGovernanceMapper {
	return NewSimpleGovernanceMapper(proposalStoreKey, optionStoreKey, false, ck, sm)
}

func NewSimpleGovernanceKeeperReadOnly(proposalStoreKey sdk.StoreKey, ck bank.CoinKeeper, sm stake.KeeperRead) SimpleGovernanceMapper {
	return NewSimpleGovernanceMapper(proposalStoreKey, optionStoreKey, true, ck, sm)
}

func NewKeeper(proposalStoreKey sdk.StoreKey, writeAccess bool, ck bank.CoinKeeper, sm stake.KeeperRead) SimpleGovernanceMapper {
	cdc = wire.NewCodec()
	return SimpleGovernanceKeeper{
		ProposalStoreKey: proposalStoreKey,
		OptionStoreKey:   optionStoreKey,
		Cdc:              cdc,
		WriteAccess:      writeAccess,
		ck:               ck,
		sm:               sm,
	}
}

func (sgm SimpleGovernanceKeeper) GetProposal(ctx sdk.Context, proposalID int64) Proposal {
	store := ctx.KVStore(sgm.ProposalStoreKey)

	bpi, err := sgm.cdc.MarshalBinary(proposalID)
	if err != nil {
		panic(error)
	}

	bp = store.Get(bpi)
	if bp == nil {
		return nil
	}

	proposal := Proposal{}

	err := sgm.cdc.UnmarshalBinary(bp, proposal)
	if err != nil {
		panic(error)
	}

	return proposal
}

func (sgm SimpleGovernanceKeeper) SetProposal(ctx sdk.Context, proposalID int64, proposal Proposal) sdk.Error {
	if !sgm.WriteAccess {
		return sdk.ErrUnauthorized("").Trace("No write access for simple governance store")
	}

	store := ctx.KVStore(sgm.ProposalStoreKey)

	bp, err := sgm.cdc.MarshalBinary(proposal)
	if err != nil {
		panic(error) // return proper error
	}

	bpi, err := sgm.cdc.MarshalBinary(proposalID)
	if err != nil {
		panic(error) // return proper error
	}

	store.set(bpi, bp)
	return nil
}

func (sgm SimpleGovernanceKeeper) NewProposalID(ctx sdk.Context) int64 {
	store := ctx.KVStore(sgm.ProposalStoreKey)

	bid := store.Get([]byte("TotalID"))
	if bid == nil {
		return 0
	}

	totalID := new(int64)
	err := sgm.cdc.UnmarshalBinary(bid, totalID)
	if err != nil {
		panic(error)
	}

	return (totalID + 1)
}

//--------------------------------------------------------------------------------------

func (sgm SimpleGovernanceKeeper) GetOption(ctx sdk.Context, key []byte) string {
	store := ctx.KVStore(sgm.proposalStoreKey)

	bv = store.Get(key)
	if bv == nil {
		return nil
	}

	option := new(string)

	err := sgm.cdc.UnmarshalBinary(bv, option)
	if err != nil {
		panic(error)
	}

	return option
}

func (sgm SimpleGovernanceKeeper) SetOption(ctx sdk.Context, key []byte, option string) sdk.Error {
	if !sgm.WriteAccess {
		return sdk.ErrUnauthorized("").Trace("No write access for simple governance store")
	}

	store := ctx.KVStore(sgm.proposalStoreKey)

	bv, err := sgm.cdc.MarshalBinary(option)
	if err != nil {
		panic(error)
	}

	store.set(key, bv)
	return nil
}

//--------------------------------------------------------------------------------------

func (sgm SimpleGovernanceKeeper) getProposalQueue(ctx sdk.Context) ProposalQueue {
	store := ctx.KVStore(gm.proposalStoreKey)
	bpq := store.Get([]byte("proposalQueue"))
	if bz == nil {
		return nil
	}

	proposalQueue := &ProposalQueue{}
	err := sgm.cdc.UnmarshalBinaryBare(bpq, proposalQueue)
	if err != nil {
		panic(err)
	}

	return proposalQueue
}

func (sgm SimpleGovernanceKeeper) setProposalQueue(ctx sdk.Context, proposalQueue ProposalQueue) {
	store := ctx.KVStore(gm.proposalStoreKey)

	bpq, err := sgm.cdc.MarshalBinaryBare(proposalQueue)
	if err != nil {
		panic(err)
	}

	store.Set([]byte("proposalQueue"), bpq)
}

func (sgm SimpleGovernanceKeeper) ProposalQueuePeek(ctx sdk.Context) Proposal {
	proposalQueue := sgm.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	return sgm.GetProposal(ctx, proposalQueue[0])
}

func (sgm SimpleGovernanceKeeper) ProposalQueuePop(ctx sdk.Context) Proposal {
	proposalQueue := sgm.getProposalQueue(ctx)
	if len(proposalQueue) == 0 {
		return nil
	}
	frontElement, proposalQueue = proposalQueue[0], proposalQueue[1:]
	sgm.setProposalQueue(ctx, proposalQueue)
	return sgm.GetProposal(ctx, frontElement)
}

func (sgm SimpleGovernanceKeeper) ProposalQueuePush(ctx sdk.Context, proposaID int64) {
	proposalQueue := append(sgm.getProposalQueue(ctx), proposalID)
	sgm.setProposalQueue(ctx, proposalQueue)
}
