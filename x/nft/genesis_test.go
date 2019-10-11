package nft_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/nft"
)

func TestInitGenesis(t *testing.T) {
	app, ctx := createTestApp(false)
	genesisState := nft.DefaultGenesisState()
	require.Equal(t, 0, len(genesisState.Owners))
	require.Equal(t, 0, len(genesisState.Collections))

	ids := []string{id, id2, id3}
	idCollection := nft.NewIDCollection(denom, ids)
	idCollection2 := nft.NewIDCollection(denom2, ids)
	owner := nft.NewOwner(address, idCollection)

	owner2 := nft.NewOwner(address2, idCollection2)

	owners := []nft.Owner{owner, owner2}

	nft1 := nft.NewBaseNFT(id, address, tokenURI1)
	nft2 := nft.NewBaseNFT(id2, address, tokenURI1)
	nft3 := nft.NewBaseNFT(id3, address, tokenURI1)
	nfts := nft.NewNFTs(&nft1, &nft2, &nft3)
	collection := nft.NewCollection(denom, nfts)

	nftx := nft.NewBaseNFT(id, address2, tokenURI1)
	nft2x := nft.NewBaseNFT(id2, address2, tokenURI1)
	nft3x := nft.NewBaseNFT(id3, address2, tokenURI1)
	nftsx := nft.NewNFTs(&nftx, &nft2x, &nft3x)
	collection2 := nft.NewCollection(denom2, nftsx)

	collections := nft.NewCollections(collection, collection2)

	genesisState = nft.NewGenesisState(owners, collections)

	nft.InitGenesis(ctx, app.NFTKeeper, genesisState)

	returnedOwners := app.NFTKeeper.GetOwners(ctx)
	require.Equal(t, 2, len(owners))
	require.Equal(t, returnedOwners[0].String(), owners[0].String())
	require.Equal(t, returnedOwners[1].String(), owners[1].String())

	returnedCollections := app.NFTKeeper.GetCollections(ctx)
	require.Equal(t, 2, len(returnedCollections))
	require.Equal(t, returnedCollections[0].String(), collections[0].String())
	require.Equal(t, returnedCollections[1].String(), collections[1].String())

	exportedGenesisState := nft.ExportGenesis(ctx, app.NFTKeeper)
	require.Equal(t, len(genesisState.Owners), len(exportedGenesisState.Owners))
	require.Equal(t, genesisState.Owners[0].String(), exportedGenesisState.Owners[0].String())
	require.Equal(t, genesisState.Owners[1].String(), exportedGenesisState.Owners[1].String())

	require.Equal(t, len(genesisState.Collections), len(exportedGenesisState.Collections))
	require.Equal(t, genesisState.Collections[0].String(), exportedGenesisState.Collections[0].String())
	require.Equal(t, genesisState.Collections[1].String(), exportedGenesisState.Collections[1].String())
}
