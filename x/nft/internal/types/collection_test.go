package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------- Collection ---------------------------------------------------

func TestNewCollection(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, tokenURI)
	nfts := NewNFTs(&testNFT, &testNFT2)
	collection := NewCollection(fmt.Sprintf("      %s      ", denom), nfts)
	require.Equal(t, collection.Denom, denom)
	require.Equal(t, len(collection.NFTs), 2)
}

func TestEmptyCollection(t *testing.T) {
	collection := EmptyCollection()
	require.Equal(t, collection.Denom, "")
	require.Equal(t, len(collection.NFTs), 0)
}

func TestCollectionGetNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	returnedNFT, err := collection.GetNFT(id)
	require.NoError(t, err)
	require.Equal(t, testNFT.String(), returnedNFT.String())

	returnedNFT, err = collection.GetNFT(id2)
	require.Error(t, err)
	require.Nil(t, returnedNFT)
}

func TestCollectionContainsNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	contains := collection.ContainsNFT(id)
	require.True(t, contains)

	contains = collection.ContainsNFT(id2)
	require.False(t, contains)
}

func TestCollectionAddNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	newCollection, err := collection.AddNFT(&testNFT)
	require.Error(t, err)
	require.Equal(t, collection.String(), newCollection.String())

	newCollection, err = collection.AddNFT(&testNFT2)
	require.NoError(t, err)
	require.NotEqual(t, collection.String(), newCollection.String())
	require.Equal(t, len(newCollection.NFTs), 2)

}

func TestCollectionUpdateNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	testNFT3 := NewBaseNFT(id, address2, tokenURI2)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	newCollection, err := collection.UpdateNFT(&testNFT2)
	require.Error(t, err)
	require.Equal(t, collection.String(), newCollection.String())

	collection, err = collection.UpdateNFT(&testNFT3)
	require.NoError(t, err)

	returnedNFT, err := collection.GetNFT(id)
	require.NoError(t, err)

	require.Equal(t, returnedNFT.GetOwner(), address2)
	require.Equal(t, returnedNFT.GetTokenURI(), tokenURI2)

}

func TestCollectionDeleteNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	testNFT3 := NewBaseNFT(id3, address, tokenURI)
	nfts := NewNFTs(&testNFT, &testNFT2)
	collection := NewCollection(denom, nfts)

	newCollection, err := collection.DeleteNFT(&testNFT3)
	require.Error(t, err)
	require.Equal(t, collection.String(), newCollection.String())

	collection, err = collection.DeleteNFT(&testNFT2)
	require.NoError(t, err)
	require.Equal(t, len(collection.NFTs), 1)

	returnedNFT, err := collection.GetNFT(id2)
	require.Nil(t, returnedNFT)
	require.Error(t, err)
}

func TestCollectionSupplyMethod(t *testing.T) {

	empty := EmptyCollection()
	require.Equal(t, empty.Supply(), 0)

	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts := NewNFTs(&testNFT, &testNFT2)
	collection := NewCollection(denom, nfts)

	require.Equal(t, collection.Supply(), 2)

	collection, err := collection.DeleteNFT(&testNFT)
	require.Nil(t, err)
	require.Equal(t, collection.Supply(), 1)

	collection, err = collection.DeleteNFT(&testNFT2)
	require.Nil(t, err)
	require.Equal(t, collection.Supply(), 0)

	collection, err = collection.AddNFT(&testNFT)
	require.Nil(t, err)
	require.Equal(t, collection.Supply(), 1)

}

func TestCollectionStringMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts := NewNFTs(&testNFT, &testNFT2)
	collection := NewCollection(denom, nfts)
	require.Equal(t, collection.String(),
		fmt.Sprintf(`Denom: 				%s
NFTs:

ID:				%s
Owner:			%s
TokenURI:		%s
ID:				%s
Owner:			%s
TokenURI:		%s`, denom, id, address.String(), tokenURI,
			id2, address2.String(), tokenURI2))
}

// ---------------------------------------- Collections ---------------------------------------------------

func TestNewCollections(t *testing.T) {

	emptyCollections := NewCollections()
	require.Empty(t, emptyCollections)

	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts2 := NewNFTs(&testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections := NewCollections(collection, collection2)
	require.Equal(t, len(collections), 2)

}
func TestCollectionsAppendMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	collections := NewCollections(collection)

	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts2 := NewNFTs(&testNFT2)
	collection2 := NewCollection(denom2, nfts2)
	collections2 := NewCollections(collection2)

	collections = collections.Append(collections2...)
	require.Equal(t, len(collections), 2)

}
func TestCollectionsFindMethod(t *testing.T) {

	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts2 := NewNFTs(&testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections := NewCollections(collection)

	foundCollection, found := collections.Find(denom2)
	require.False(t, found)
	require.Empty(t, foundCollection)

	collections = NewCollections(collection, collection2)

	foundCollection, found = collections.Find(denom2)
	require.True(t, found)
	require.Equal(t, foundCollection.String(), collection2.String())

	collection3 := NewCollection(denom3, nfts)
	collections = NewCollections(collection, collection2, collection3)

	_, found = collections.Find(denom)
	require.True(t, found)

	_, found = collections.Find(denom2)
	require.True(t, found)

	_, found = collections.Find(denom3)
	require.True(t, found)
}

func TestCollectionsRemoveMethod(t *testing.T) {

	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	collections := NewCollections(collection)

	returnedCollections, removed := collections.Remove(denom2)
	require.False(t, removed)
	require.Equal(t, returnedCollections.String(), collections.String())

	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts2 := NewNFTs(&testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections = NewCollections(collection, collection2)

	returnedCollections, removed = collections.Remove(denom2)
	require.True(t, removed)
	require.NotEqual(t, returnedCollections.String(), collections.String())
	require.Equal(t, 1, len(returnedCollections))

	foundCollection, found := returnedCollections.Find(denom2)
	require.False(t, found)
	require.Empty(t, foundCollection)
}

func TestCollectionsStringMethod(t *testing.T) {
	collections := NewCollections()
	require.Equal(t, collections.String(), "")

	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts2 := NewNFTs(&testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections = NewCollections(collection, collection2)
	require.Equal(t, fmt.Sprintf(`Denom: 				%s
NFTs:

ID:				%s
Owner:			%s
TokenURI:		%s
Denom: 				%s
NFTs:

ID:				%s
Owner:			%s
TokenURI:		%s`, denom, id, address.String(), tokenURI,
		denom2, id2, address2.String(), tokenURI2), collections.String())

}

func TestCollectionsEmptyMethod(t *testing.T) {

	collections := NewCollections()
	require.True(t, collections.Empty())

	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	collections = NewCollections(collection)
	require.False(t, collections.Empty())

}

func TestCollectionsSortInterface(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts2 := NewNFTs(&testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections := NewCollections(collection, collection2)
	require.Equal(t, 2, collections.Len())

	require.True(t, collections.Less(0, 1))
	require.False(t, collections.Less(1, 0))

	collections.Swap(0, 1)
	require.False(t, collections.Less(0, 1))
	require.True(t, collections.Less(1, 0))

	collections.Sort()
	require.True(t, collections.Less(0, 1))
	require.False(t, collections.Less(1, 0))
}

func TestCollectionMarshalAndUnmarshalJSON(t *testing.T) {
	testNFT := NewBaseNFT(id, address, tokenURI)
	nfts := NewNFTs(&testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, address2, tokenURI2)
	nfts2 := NewNFTs(&testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections := NewCollections(collection, collection2)

	bz, err := collections.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, string(bz), fmt.Sprintf(`{"%s":{"nfts":{"%s":{"id":"%s","owner":"%s","token_uri":"%s"}}},"%s":{"nfts":{"%s":{"id":"%s","owner":"%s","token_uri":"%s"}}}}`,
		denom, id, id, address.String(), tokenURI,
		denom2, id2, id2, address2.String(), tokenURI2,
	))

	var newCollections Collections
	err = newCollections.UnmarshalJSON(bz)
	require.NoError(t, err)

	err = newCollections.UnmarshalJSON([]byte{})
	require.Error(t, err)
}
