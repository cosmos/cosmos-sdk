package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------- BaseNFT ---------------------------------------------------

func TestBaseNFTGetMethods(t *testing.T) {

	testNFT := NewBaseNFT(id, address, tokenURI)

	require.Equal(t, id, testNFT.GetID())
	require.Equal(t, address, testNFT.GetOwner())
	require.Equal(t, tokenURI, testNFT.GetTokenURI())
}

func TestBaseNFTSetMethods(t *testing.T) {

	testNFT := NewBaseNFT(id, address, tokenURI)

	testNFT.SetOwner(address2)
	require.Equal(t, address2, testNFT.GetOwner())

	testNFT.EditMetadata(tokenURI2)
	require.Equal(t, tokenURI2, testNFT.GetTokenURI())
}

func TestBaseNFTStringFormat(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	expected := fmt.Sprintf(`ID:				%s
Owner:			%s
TokenURI:		%s`,
		id, address, tokenURI)
	require.Equal(t, expected, testNFT.String())
}

// ---------------------------------------- NFTs ---------------------------------------------------

func TestNewNFTs(t *testing.T) {

	emptyNFTs := NewNFTs()
	require.Equal(t, len(emptyNFTs), 0)

	testNFT := NewBaseNFT(id, address, tokenURI)
	oneNFTs := NewNFTs(&testNFT)
	require.Equal(t, len(oneNFTs), 1)

	testNFT2 := NewBaseNFT(id2, address, tokenURI)
	twoNFTs := NewNFTs(&testNFT, &testNFT2)
	require.Equal(t, len(twoNFTs), 2)

}

func TestNFTsAppendMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	require.Equal(t, len(nfts), 1)

	testNFT2 := NewBaseNFT(id2, address, tokenURI)
	nfts2 := NewNFTs(&testNFT2)

	nfts = nfts.Append(nfts2...)
	require.Equal(t, len(nfts), 2)

	var id3 = string('3')
	var id4 = string('4')
	var id5 = string('5')
	testNFT3 := NewBaseNFT(id3, address, tokenURI)
	testNFT4 := NewBaseNFT(id4, address, tokenURI)
	testNFT5 := NewBaseNFT(id5, address, tokenURI)

	nfts3 := NewNFTs(&testNFT5, &testNFT3, &testNFT4)
	nfts = nfts.Append(nfts3...)
	require.Equal(t, len(nfts), 5)

	nft, found := nfts.Find(id2)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT2.String())

	nft, found = nfts.Find(id5)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT5.String())

	nft, found = nfts.Find(id3)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT3.String())
}

func TestNFTsFindMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, tokenURI)

	var id3 = string('3')
	var id4 = string('4')
	var id5 = string('5')
	testNFT3 := NewBaseNFT(id3, address, tokenURI)
	testNFT4 := NewBaseNFT(id4, address, tokenURI)
	testNFT5 := NewBaseNFT(id5, address, tokenURI)

	nfts := NewNFTs(&testNFT, &testNFT3, &testNFT4, &testNFT5, &testNFT2)
	nft, found := nfts.Find(id)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT.String())

	nft, found = nfts.Find(id2)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT2.String())

	nft, found = nfts.Find(id3)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT3.String())

	nft, found = nfts.Find(id4)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT4.String())

	nft, found = nfts.Find(id5)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT5.String())

	var id6 = string('6')
	nft, found = nfts.Find(id6)
	require.False(t, found)
	require.Nil(t, nft)
}

func TestNFTsUpdateMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	var success bool
	nfts, success = nfts.Update(id, &testNFT2)
	require.True(t, success)

	nft, found := nfts.Find(id2)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT2.String())

	nft, found = nfts.Find(id)
	require.False(t, found)
	require.Nil(t, nft)

	var returnedNFTs NFTs
	returnedNFTs, success = nfts.Update(id, &testNFT2)
	require.False(t, success)
	require.Equal(t, returnedNFTs.String(), nfts.String())

}

func TestNFTsRemoveMethod(t *testing.T) {

	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, tokenURI)
	nfts := NewNFTs(&testNFT, &testNFT2)

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
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	require.Equal(t, nfts.String(), fmt.Sprintf(`ID:				%s
Owner:			%s
TokenURI:		%s`, id, address, tokenURI))
}

func TestNFTsEmptyMethod(t *testing.T) {
	nfts := NewNFTs()
	require.True(t, nfts.Empty())
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts = NewNFTs(&testNFT)
	require.False(t, nfts.Empty())
}

func TestNFTsMarshalUnmarshalJSON(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	bz, err := nfts.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz),
		fmt.Sprintf(`{"%s":{"id":"%s","owner":"%s","token_uri":"%s"}}`,
			id, id, address.String(), tokenURI))

	var unmarshaledNFTs NFTs
	err = unmarshaledNFTs.UnmarshalJSON(bz)
	require.NoError(t, err)
	require.Equal(t, unmarshaledNFTs.String(), nfts.String())

	bz = []byte{}
	err = unmarshaledNFTs.UnmarshalJSON(bz)
	require.Error(t, err)
}

func TestNFTsSortInterface(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, tokenURI)

	nfts := NewNFTs(&testNFT)
	require.Equal(t, nfts.Len(), 1)

	nfts = NewNFTs(&testNFT, &testNFT2)
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
