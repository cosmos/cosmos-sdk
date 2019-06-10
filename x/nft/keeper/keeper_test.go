package keeper

import (
	"bytes"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/nft/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
)

// for incode address generation
func testAddr(addr string, bech string) sdk.AccAddress {

	res, err := sdk.AccAddressFromHex(addr)
	if err != nil {
		panic(err)
	}
	bechexpected := res.String()
	if bech != bechexpected {
		panic("Bech encoding doesn't match reference")
	}

	bechres, err := sdk.AccAddressFromBech32(bech)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(bechres, res) {
		panic("Bech decode and hex decode don't match")
	}

	return res
}

// nolint: unparam
func createTestAddrs(numAddrs int) []sdk.AccAddress {
	var addresses []sdk.AccAddress
	var buffer bytes.Buffer

	// start at 100 so we can make up to 999 test addresses with valid test addresses
	for i := 100; i < (numAddrs + 100); i++ {
		numString := strconv.Itoa(i)
		buffer.WriteString("A58856F0FD53BF058B4909A21AEC019107BA6") //base address string

		buffer.WriteString(numString) //adding on final two digits to make addresses unique
		res, _ := sdk.AccAddressFromHex(buffer.String())
		bech := res.String()
		addresses = append(addresses, testAddr(buffer.String(), bech))
		buffer.Reset()
	}
	return addresses
}

func initialize() (ctx sdk.Context, keeperInstance Keeper) {
	cdc := codec.New()
	types.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	keyNFT := sdk.NewKVStoreKey(types.StoreKey)
	keeperInstance = NewKeeper(cdc, keyNFT)

	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(keyNFT, sdk.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	if err != nil {
		panic(err)
	}
	ctx = sdk.NewContext(ms, abci.Header{ChainID: "test-chain"}, true, log.NewNopLogger())
	ctx = ctx.WithConsensusParams(
		&abci.ConsensusParams{
			Validator: &abci.ValidatorParams{
				PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519},
			},
		},
	)
	return
}

func TestMintNFT(t *testing.T) {
	addresses := createTestAddrs(1)
	ctx, keeper := initialize()

	denom := sdk.DefaultBondDenom
	id1 := "1"
	address1 := addresses[0]
	tokenURI1 := "https://google.com"
	description1 := "test_description"
	image1 := "test_image"
	name1 := "test_name"

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id1, address1, tokenURI1, description1, image1, name1)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	id2 := "2"

	// MintNFT shouldn't fail when collection exists
	nft2 := types.NewBaseNFT(id2, address1, tokenURI1, description1, image1, name1)
	err = keeper.MintNFT(ctx, denom, nft2)
	require.Nil(t, err)
}

func TestGetNFT(t *testing.T) {
	addresses := createTestAddrs(1)
	ctx, keeper := initialize()

	denom := sdk.DefaultBondDenom
	id1 := "1"
	address1 := addresses[0]
	tokenURI1 := "https://google.com"
	description1 := "test_description"
	image1 := "test_image"
	name1 := "test_name"

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id1, address1, tokenURI1, description1, image1, name1)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// GetNFT should get the NFT
	var receivedNFT types.NFT
	receivedNFT, err = keeper.GetNFT(ctx, denom, id1)
	require.Nil(t, err)
	require.Equal(t, receivedNFT.GetID(), id1)
	require.True(t, receivedNFT.GetOwner().Equals(address1))
	require.Equal(t, receivedNFT.GetTokenURI(), tokenURI1)
	require.Equal(t, receivedNFT.GetDescription(), description1)
	require.Equal(t, receivedNFT.GetImage(), image1)
	require.Equal(t, receivedNFT.GetName(), name1)

	id2 := "2"

	// MintNFT shouldn't fail when collection exists
	nft2 := types.NewBaseNFT(id2, address1, tokenURI1, description1, image1, name1)
	err = keeper.MintNFT(ctx, denom, nft2)
	require.Nil(t, err)

	// GetNFT should get the NFT when collection exists
	var receivedNFT2 types.NFT
	receivedNFT2, err = keeper.GetNFT(ctx, denom, id2)
	require.Nil(t, err)
	require.Equal(t, receivedNFT2.GetID(), id2)
	require.True(t, receivedNFT2.GetOwner().Equals(address1))
	require.Equal(t, receivedNFT2.GetTokenURI(), tokenURI1)
	require.Equal(t, receivedNFT2.GetDescription(), description1)
	require.Equal(t, receivedNFT2.GetImage(), image1)
	require.Equal(t, receivedNFT2.GetName(), name1)

}

func TestUpdateNFT(t *testing.T) {
	addresses := createTestAddrs(1)
	ctx, keeper := initialize()

	denom := sdk.DefaultBondDenom
	id1 := "1"
	address1 := addresses[0]
	tokenURI1 := "https://google.com"
	description1 := "test_description"
	image1 := "test_image"
	name1 := "test_name"

	nft := types.NewBaseNFT(id1, address1, tokenURI1, description1, image1, name1)

	// UpdateNFT should fail when NFT doesn't exists
	err := keeper.UpdateNFT(ctx, denom, nft)
	require.NotNil(t, err)

	// MintNFT shouldn't fail when collection does not exist
	err = keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	tokenURI2 := "https://facebook.com"

	// UpdateNFT shouldn't fail when NFT exists
	nft2 := types.NewBaseNFT(id1, address1, tokenURI2, description1, image1, name1)
	err = keeper.UpdateNFT(ctx, denom, nft2)
	require.Nil(t, err)

	// GetNFT should get the NFT with new tokenURI
	var receivedNFT types.NFT
	receivedNFT, err = keeper.GetNFT(ctx, denom, id1)
	require.Nil(t, err)
	require.Equal(t, receivedNFT.GetTokenURI(), tokenURI2)

}

func TestDeleteNFT(t *testing.T) {
	addresses := createTestAddrs(1)
	ctx, keeper := initialize()

	denom := sdk.DefaultBondDenom
	id1 := "1"
	address1 := addresses[0]
	tokenURI1 := "https://google.com"
	description1 := "test_description"
	image1 := "test_image"
	name1 := "test_name"

	// DeleteNFT should fail when NFT doesn't exist and collection doesn't exist
	err := keeper.DeleteNFT(ctx, denom, id1)
	require.NotNil(t, err)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id1, address1, tokenURI1, description1, image1, name1)
	err = keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	id2 := "2"

	// DeleteNFT should fail when NFT doesn't exist but collection does exist
	err = keeper.DeleteNFT(ctx, denom, id2)
	require.NotNil(t, err)

	// DeleteNFT should not fail when NFT and collection exist
	err = keeper.DeleteNFT(ctx, denom, id1)
	require.Nil(t, err)

	// NFT should no longer exist
	isNFT := keeper.IsNFT(ctx, denom, id1)
	require.False(t, isNFT)
}

func TestIsNFT(t *testing.T) {
	addresses := createTestAddrs(1)
	ctx, keeper := initialize()

	denom := sdk.DefaultBondDenom
	id1 := "1"
	address1 := addresses[0]
	tokenURI1 := "https://google.com"
	description1 := "test_description"
	image1 := "test_image"
	name1 := "test_name"

	// IsNFT should return false
	isNFT := keeper.IsNFT(ctx, denom, id1)
	require.False(t, isNFT)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id1, address1, tokenURI1, description1, image1, name1)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// IsNFT should return true
	isNFT = keeper.IsNFT(ctx, denom, id1)
	require.True(t, isNFT)
}

func TestSetCollection(t *testing.T) {
	addresses := createTestAddrs(1)
	ctx, keeper := initialize()

	denom := sdk.DefaultBondDenom
	id1 := "1"
	address1 := addresses[0]
	tokenURI1 := "https://google.com"
	description1 := "test_description"
	image1 := "test_image"
	name1 := "test_name"

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id1, address1, tokenURI1, description1, image1, name1)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// collection should exist
	collection, exists := keeper.GetCollection(ctx, denom)
	require.True(t, exists)

	id2 := "2"
	nft2 := types.NewBaseNFT(id2, address1, tokenURI1, description1, image1, name1)
	collection.AddNFT(nft2)
	keeper.SetCollection(ctx, denom, collection)

	collection, exists = keeper.GetCollection(ctx, denom)
	require.True(t, exists)
	require.Len(t, collection.NFTs, 2)

}
func TestGetCollection(t *testing.T) {
	addresses := createTestAddrs(1)
	ctx, keeper := initialize()

	denom := sdk.DefaultBondDenom
	id1 := "1"
	address1 := addresses[0]
	tokenURI1 := "https://google.com"
	description1 := "test_description"
	image1 := "test_image"
	name1 := "test_name"

	// collection shouldn't exist
	collection, exists := keeper.GetCollection(ctx, denom)
	require.Empty(t, collection)
	require.False(t, exists)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id1, address1, tokenURI1, description1, image1, name1)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// collection should exist
	collection, exists = keeper.GetCollection(ctx, denom)
	require.True(t, exists)
	require.NotEmpty(t, collection)
}
func TestGetCollections(t *testing.T) {
	addresses := createTestAddrs(1)
	ctx, keeper := initialize()

	denom := sdk.DefaultBondDenom
	id1 := "1"
	address1 := addresses[0]
	tokenURI1 := "https://google.com"
	description1 := "test_description"
	image1 := "test_image"
	name1 := "test_name"

	// collections should be empty
	collections := keeper.GetCollections(ctx)
	require.Empty(t, collections)

	// MintNFT shouldn't fail when collection does not exist
	nft := types.NewBaseNFT(id1, address1, tokenURI1, description1, image1, name1)
	err := keeper.MintNFT(ctx, denom, nft)
	require.Nil(t, err)

	// collections should equal 1
	collections = keeper.GetCollections(ctx)
	require.NotEmpty(t, collections)
	require.Equal(t, len(collections), 1)
}
