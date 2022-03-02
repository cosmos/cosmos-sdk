package group

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group/keeper"
	"github.com/regen-network/regen-ledger/x/group"
)

func EndBlocker(ctx sdk.Context, k keeper.Keeper) {
	storeKey := sdk.NewKVStoreKey(StoreKey)
	timeBytes := sdk.FormatTimeBytes(ctx.BlockTime())
	it, _ := k.ProposalsByVotingPeriodEnd.Get(ctx.KVStore(storeKey), sdk.PrefixEndBytes(timeBytes))

	var proposal group.Proposal
	for {
		_, err = it.LoadNext(&proposal)
		if errors.ErrORMIteratorDone.Is(err) {
			break
		}
	}
}
