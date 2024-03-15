package keeper

import (
	"context"
	"sort"

	"cosmossdk.io/x/nft"
)

// InitGenesis initializes the nft module's genesis state from a given
// genesis state.
func (k Keeper) InitGenesis(ctx context.Context, data *nft.GenesisState) error {
	for _, class := range data.Classes {
		if err := k.SaveClass(ctx, *class); err != nil {
			return err
		}
	}
	for _, entry := range data.Entries {
		for _, nft := range entry.Nfts {
			owner, err := k.ac.StringToBytes(entry.Owner)
			if err != nil {
				return err
			}

			if err := k.Mint(ctx, *nft, owner); err != nil {
				return err
			}
		}
	}
	return nil
}

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx context.Context) (*nft.GenesisState, error) {
	classes := k.GetClasses(ctx)
	nftMap := make(map[string][]*nft.NFT)
	for _, class := range classes {
		nfts := k.GetNFTsOfClass(ctx, class.Id)
		for i, n := range nfts {
			owner := k.GetOwner(ctx, n.ClassId, n.Id)
			ownerStr, err := k.ac.BytesToString(owner.Bytes())
			if err != nil {
				return nil, err
			}
			nftArr, ok := nftMap[ownerStr]
			if !ok {
				nftArr = make([]*nft.NFT, 0)
			}
			nftMap[ownerStr] = append(nftArr, &nfts[i])
		}
	}

	owners := make([]string, 0, len(nftMap))
	for owner := range nftMap {
		owners = append(owners, owner)
	}
	sort.Strings(owners)

	entries := make([]*nft.Entry, 0, len(nftMap))
	for _, owner := range owners {
		entries = append(entries, &nft.Entry{
			Owner: owner,
			Nfts:  nftMap[owner],
		})
	}
	return &nft.GenesisState{
		Classes: classes,
		Entries: entries,
	}, nil
}
