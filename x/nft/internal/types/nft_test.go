package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------- BaseNFT ---------------------------------------------------

func TestBaseNFTGetMethods(t *testing.T) {

	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)

	require.Equal(t, id, testNFT.GetID())
	require.Equal(t, address, testNFT.GetOwner())
	require.Equal(t, name, testNFT.GetName())
	require.Equal(t, description, testNFT.GetDescription())
	require.Equal(t, image, testNFT.GetImage())
	require.Equal(t, tokenURI, testNFT.GetTokenURI())
}

func TestBaseNFTSetMethods(t *testing.T) {

	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)

	testNFT = testNFT.SetOwner(address2)
	require.Equal(t, address2, testNFT.GetOwner())

	testNFT = testNFT.EditMetadata(name2, description2, image2, tokenURI2)
	require.Equal(t, name2, testNFT.GetName())
	require.Equal(t, description2, testNFT.GetDescription())
	require.Equal(t, image2, testNFT.GetImage())
	require.Equal(t, tokenURI2, testNFT.GetTokenURI())
}

func TestBaseNFTStringFormat(t *testing.T) {
	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	expected := fmt.Sprintf(`ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`,
		id, address, name, description, image, tokenURI)
	require.Equal(t, expected, testNFT.String())
}

// ---------------------------------------- NFTs ---------------------------------------------------

func TestNewNFTs(t *testing.T) {

	emptyNFTs := NewNFTs()
	require.Equal(t, len(emptyNFTs), 0)

	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	oneNFTs := NewNFTs(testNFT)
	require.Equal(t, len(oneNFTs), 1)

	testNFT2 := NewBaseNFT(id2, address, name, description, image, tokenURI)
	twoNFTs := NewNFTs(testNFT, testNFT2)
	require.Equal(t, len(twoNFTs), 2)

}

func TestNFTsAddMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	require.Equal(t, len(nfts), 1)

	testNFT2 := NewBaseNFT(id2, address, name, description, image, tokenURI)
	nfts2 := NewNFTs(testNFT2)

	nfts = nfts.Add(nfts2)
	require.Equal(t, len(nfts), 2)
}

func TestNFTsFindMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT, testNFT2)

	nft, found := nfts.Find(id)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT.String())

	nft, found = nfts.Find(id3)
	require.False(t, found)
	require.Nil(t, nft)
}

func TestNFTsUpdateMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	var success bool
	nfts, success = nfts.Update(id, testNFT2)
	require.True(t, success)

	nft, found := nfts.Find(id2)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT2.String())

	nft, found = nfts.Find(id)
	require.False(t, found)
	require.Nil(t, nft)

	var returnedNFTs NFTs
	returnedNFTs, success = nfts.Update(id, testNFT2)
	require.False(t, success)
	require.Equal(t, returnedNFTs.String(), nfts.String())

}

func TestNFTsRemoveMethod(t *testing.T) {

	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT, testNFT2)

	var success bool
	nfts, success = nfts.Remove(id)
	require.True(t, success)
	require.Equal(t, len(nfts), 1)

	nfts, success = nfts.Remove(id2)
	require.True(t, success)
	require.Equal(t, len(nfts), 0)

	var returnedNFTs NFTs
	returnedNFTs, success = nfts.Remove(id2)
	require.False(t, success)
	require.Equal(t, nfts.String(), returnedNFTs.String())
}

func TestNFTsStringMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	require.Equal(t, nfts.String(), fmt.Sprintf(`ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`, id, address, name, description, image, tokenURI))
}

func TestNFTsEmptyMethod(t *testing.T) {
	nfts := NewNFTs()
	require.True(t, nfts.Empty())
	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	nfts = NewNFTs(testNFT)
	require.False(t, nfts.Empty())
}

func TestNFTsMarshalUnmarshalJSON(t *testing.T) {
	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	bz, err := nfts.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, string(bz),
		fmt.Sprintf(`{"%s":{"id":"%s","owner":"%s","name":"%s","description":"%s","image":"%s","token_uri":"%s"}}`,
			id, id, address.String(), name, description, image, tokenURI))

	var unmarshaledNFTs NFTs
	err = unmarshaledNFTs.UnmarshalJSON(bz)
	require.Nil(t, err)
	require.Equal(t, unmarshaledNFTs.String(), nfts.String())

	bz = []byte{}
	err = unmarshaledNFTs.UnmarshalJSON(bz)
	require.NotNil(t, err)
}

func TestNFTsSortInterface(t *testing.T) {
	testNFT := NewBaseNFT(id, address, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, name, description, image, tokenURI)

	nfts := NewNFTs(testNFT)
	require.Equal(t, nfts.Len(), 1)

	nfts = NewNFTs(testNFT, testNFT2)
	require.Equal(t, nfts.Len(), 2)

	require.True(t, nfts.Less(0, 1))
	require.False(t, nfts.Less(1, 0))

	nfts.Swap(0, 1)
	require.False(t, nfts.Less(0, 1))
	require.True(t, nfts.Less(1, 0))

	nfts.Sort()
	require.True(t, nfts.Less(0, 1))
	require.False(t, nfts.Less(1, 0))
}
