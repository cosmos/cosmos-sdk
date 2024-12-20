package simulation

import (
	"context"
	"github.com/cosmos/cosmos-sdk/simsx/common"
	"github.com/cosmos/cosmos-sdk/simsx/module"

	"cosmossdk.io/x/nft"
	"cosmossdk.io/x/nft/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func MsgSendFactory(k keeper.Keeper) module.SimMsgFactoryFn[*nft.MsgSend] {
	return func(ctx context.Context, testData *common.ChainDataSource, reporter common.SimulationReporter) ([]common.SimAccount, *nft.MsgSend) {
		from := testData.AnyAccount(reporter, common.WithSpendableBalance())
		to := testData.AnyAccount(reporter, common.ExcludeAccounts(from))
		if reporter.IsAborted() {
			return nil, nil
		}
		n, err := randNFT(ctx, testData.Rand(), k, from.Address)
		if err != nil {
			reporter.Skip(err.Error())
			return nil, nil
		}
		msg := &nft.MsgSend{
			ClassId:  n.ClassId,
			Id:       n.Id,
			Sender:   from.AddressBech32,
			Receiver: to.AddressBech32,
		}

		return []common.SimAccount{from}, msg
	}
}

// randNFT picks a random NFT from a class belonging to the specified owner(minter).
func randNFT(ctx context.Context, r *common.XRand, k keeper.Keeper, minter sdk.AccAddress) (nft.NFT, error) {
	c, err := randClass(ctx, r, k)
	if err != nil {
		return nft.NFT{}, err
	}

	if ns := k.GetNFTsOfClassByOwner(ctx, c.Id, minter); len(ns) > 0 {
		return ns[r.Intn(len(ns))], nil
	}

	n := nft.NFT{
		ClassId: c.Id,
		Id:      r.StringN(10),
		Uri:     r.StringN(10),
	}
	return n, k.Mint(ctx, n, minter)
}

// randClass picks a random Class.
func randClass(ctx context.Context, r *common.XRand, k keeper.Keeper) (nft.Class, error) {
	if classes := k.GetClasses(ctx); len(classes) != 0 {
		return *classes[r.Intn(len(classes))], nil
	}
	c := nft.Class{
		Id:          r.StringN(10),
		Name:        r.StringN(10),
		Symbol:      r.StringN(10),
		Description: r.StringN(10),
		Uri:         r.StringN(10),
	}
	return c, k.SaveClass(ctx, c)
}
