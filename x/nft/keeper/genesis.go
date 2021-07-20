package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

// InitGenesis new nft genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *nft.GenesisState) {
	for _, class := range data.Classes {
		if err := k.NewClass(ctx, class); err != nil {
			panic(err)
		}

	}
	for _, entry := range data.Entries {
		for _, nft := range entry.NFTs {
			owner, err := sdk.AccAddressFromBech32(entry.Owner)
			if err != nil {
				panic(err)
			}

			if err := k.Mint(ctx, nft, owner); err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *nft.GenesisState {
	classes := k.GetClasses(ctx)
	nftMap := make(map[string][]nft.NFT)
	for _, class := range classes {
		nfts := k.GetNFTsOfClass(ctx, class.Id)
		for _, n := range nfts {
			owner := k.GetOwner(ctx, n.ClassId, n.Id)
			nftArr, ok := nftMap[owner.String()]
			if !ok {
				nftArr = make([]nft.NFT, 0)
			}
			nftMap[owner.String()] = append(nftArr, n)
		}
	}

	entries := make([]nft.Entry, 0, len(nftMap))
	for k, v := range nftMap {
		entries = append(entries, nft.Entry{
			Owner: k,
			NFTs:  v,
		})
	}
	return &nft.GenesisState{
		Classes: classes,
		Entries: entries,
	}
}
