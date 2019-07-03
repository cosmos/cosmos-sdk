package nft

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/x/nft/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/nft/internal/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	ctx, k, _ := keeper.Initialize()
	genesisState := DefaultGenesisState()
	require.Equal(t, 0, len(genesisState.Owners))
	require.Equal(t, 0, len(genesisState.Collections))

	ids := []string{id, id2, id3}
	idCollection := NewIDCollection(denom, ids)
	idCollection2 := NewIDCollection(denom2, ids)
	owner := types.NewOwner(address, idCollection)

	owner2 := types.NewOwner(address2, idCollection2)

	owners := []types.Owner{owner, owner2}

	nft := types.NewBaseNFT(id, address, name, description, image, tokenURI)
	nft2 := types.NewBaseNFT(id2, address, name, description, image, tokenURI)
	nft3 := types.NewBaseNFT(id3, address, name, description, image, tokenURI)
	nfts := types.NewNFTs(nft, nft2, nft3)
	collection := types.NewCollection(denom, nfts)

	nftx := types.NewBaseNFT(id, address2, name, description, image, tokenURI)
	nft2x := types.NewBaseNFT(id2, address2, name, description, image, tokenURI)
	nft3x := types.NewBaseNFT(id3, address2, name, description, image, tokenURI)
	nftsx := types.NewNFTs(nftx, nft2x, nft3x)
	collection2 := types.NewCollection(denom2, nftsx)

	collections := types.NewCollections(collection, collection2)

	genesisState = NewGenesisState(owners, collections)

	InitGenesis(ctx, k, genesisState)

	returnedOwners := k.GetOwners(ctx)
	require.Equal(t, 2, len(owners))
	require.Equal(t, returnedOwners[0].String(), owners[0].String())
	require.Equal(t, returnedOwners[1].String(), owners[1].String())

	returnedCollections := k.GetCollections(ctx)
	require.Equal(t, 2, len(returnedCollections))
	require.Equal(t, returnedCollections[0].String(), collections[0].String())
	require.Equal(t, returnedCollections[1].String(), collections[1].String())

	exportedGenesisState := ExportGenesis(ctx, k)
	require.Equal(t, len(genesisState.Owners), len(exportedGenesisState.Owners))
	require.Equal(t, genesisState.Owners[0].String(), exportedGenesisState.Owners[0].String())
	require.Equal(t, genesisState.Owners[1].String(), exportedGenesisState.Owners[1].String())

	require.Equal(t, len(genesisState.Collections), len(exportedGenesisState.Collections))
	require.Equal(t, genesisState.Collections[0].String(), exportedGenesisState.Collections[0].String())
	require.Equal(t, genesisState.Collections[1].String(), exportedGenesisState.Collections[1].String())
}
