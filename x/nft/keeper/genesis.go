package keeper

import (
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/nft"
)

// InitGenesis new nft genesis
func (k Keeper) InitGenesis(ctx sdk.Context, data *nft.GenesisState) {
	for _, class := range data.Classes {
		if err := k.SaveClass(ctx, *class); err != nil {
			panic(err)
		}
	}
	for _, entry := range data.Entries {
		for _, nft := range entry.Nfts {
			owner := sdk.MustAccAddressFromBech32(entry.Owner)

			if err := k.Mint(ctx, *nft, owner); err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx sdk.Context) *nft.GenesisState {
	classes := k.GetClasses(ctx)
	nftMap := make(map[string][]*nft.NFT)
	for _, class := range classes {
		nfts := k.GetNFTsOfClass(ctx, class.Id)
		for i, n := range nfts {
			owner := k.GetOwner(ctx, n.ClassId, n.Id)
			nftArr, ok := nftMap[owner.String()]
			if !ok {
				nftArr = make([]*nft.NFT, 0)
			}
			nftMap[owner.String()] = append(nftArr, &nfts[i])
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
	}
}

// InitGenesisFrom new nft genesis from given file path
func (k Keeper) InitGenesisFrom(ctx sdk.Context, cdc codec.JSONCodec, importPath string) error {
	f, err := module.OpenGenesisModuleFile(importPath, nft.ModuleName)
	if err != nil {
		return err
	}
	defer f.Close()

	bz, err := module.FileRead(f)
	if err != nil {
		return err
	}

	var gs nft.GenesisState
	cdc.MustUnmarshalJSON(bz, &gs)
	k.InitGenesis(ctx, &gs)
	return nil
}

// ExportGenesisTo exports a GenesisState for a given context to a given file path.
func (k Keeper) ExportGenesisTo(ctx sdk.Context, cdc codec.JSONCodec, exportPath string) error {
	f, err := module.CreateGenesisExportFile(exportPath, nft.ModuleName)
	if err != nil {
		return err
	}
	defer f.Close()

	gs := k.ExportGenesis(ctx)
	bz := cdc.MustMarshalJSON(gs)
	return module.FileWrite(f, bz)
}
