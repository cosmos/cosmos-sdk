package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	denom         = "test-denom"
	denom2        = "test-denom2"
	denom3        = "test-denom3"
	id            = "1"
	id2           = "2"
	id3           = "3"
	ownerAddress  = userAddr1
	ownerAddress2 = userAddr2
	name          = "cool token"
	name2         = "cooler token"
	description   = "a very cool token"
	description2  = "a super cool token"
	image         = "https://google.com/token-1.png"
	image2        = "https://google.com/token-2.png"
	tokenURI      = "https://google.com/token-1.json"
	tokenURI2     = "https://google.com/token-2.json"
)

// ---------------------------------------- BaseNFT ---------------------------------------------------

func TestBaseNFTGetMethods(t *testing.T) {

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)

	require.Equal(t, id, testNFT.GetID())
	require.Equal(t, ownerAddress, testNFT.GetOwner())
	require.Equal(t, name, testNFT.GetName())
	require.Equal(t, description, testNFT.GetDescription())
	require.Equal(t, image, testNFT.GetImage())
	require.Equal(t, tokenURI, testNFT.GetTokenURI())
}

func TestBaseNFTSetMethods(t *testing.T) {

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)

	testNFT = testNFT.SetOwner(ownerAddress2)
	require.Equal(t, ownerAddress2, testNFT.GetOwner())

	testNFT = testNFT.EditMetadata(name2, description2, image2, tokenURI2)
	require.Equal(t, name2, testNFT.GetName())
	require.Equal(t, description2, testNFT.GetDescription())
	require.Equal(t, image2, testNFT.GetImage())
	require.Equal(t, tokenURI2, testNFT.GetTokenURI())
}

func TestBaseNFTStringFormat(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	expected := fmt.Sprintf(`ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`,
		id, ownerAddress, name, description, image, tokenURI)
	require.Equal(t, expected, testNFT.String())
}

// ---------------------------------------- NFTs ---------------------------------------------------

func TestNewNFTs(t *testing.T) {

	emptyNFTs := NewNFTs()
	require.Equal(t, len(emptyNFTs), 0)

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	oneNFTs := NewNFTs(testNFT)
	require.Equal(t, len(oneNFTs), 1)

	testNFT2 := NewBaseNFT(id2, ownerAddress, name, description, image, tokenURI)
	twoNFTs := NewNFTs(testNFT, testNFT2)
	require.Equal(t, len(twoNFTs), 2)

}

func TestNFTsAddMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	require.Equal(t, len(nfts), 1)

	testNFT2 := NewBaseNFT(id2, ownerAddress, name, description, image, tokenURI)
	nfts2 := NewNFTs(testNFT2)

	nfts = nfts.Add(nfts2)
	require.Equal(t, len(nfts), 2)
}

func TestNFTsFindMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT, testNFT2)

	nft, found := nfts.Find(id)
	require.True(t, found)
	require.Equal(t, nft.String(), testNFT.String())

	nft, found = nfts.Find(id3)
	require.False(t, found)
	require.Nil(t, nft)
}

func TestNFTsUpdateMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress, name, description, image, tokenURI)
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

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress, name, description, image, tokenURI)
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
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	require.Equal(t, nfts.String(), fmt.Sprintf(`ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`, id, ownerAddress, name, description, image, tokenURI))
}

func TestNFTsEmptyMethod(t *testing.T) {
	nfts := NewNFTs()
	require.True(t, nfts.Empty())
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts = NewNFTs(testNFT)
	require.False(t, nfts.Empty())
}

func TestNFTsMarshalUnmarshalJSON(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	bz, err := nfts.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, string(bz),
		fmt.Sprintf(`{"%s":{"id":"%s","owner":"%s","name":"%s","description":"%s","image":"%s","token_uri":"%s"}}`,
			id, id, ownerAddress.String(), name, description, image, tokenURI))

	var unmarshaledNFTs NFTs
	err = unmarshaledNFTs.UnmarshalJSON(bz)
	require.Nil(t, err)
	require.Equal(t, unmarshaledNFTs.String(), nfts.String())

	bz = []byte{}
	err = unmarshaledNFTs.UnmarshalJSON(bz)
	require.NotNil(t, err)
}

func TestNFTsSortInterface(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress, name, description, image, tokenURI)

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

// ---------------------------------------- Collection ---------------------------------------------------

func TestNewCollection(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT, testNFT2)
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
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	returnedNFT, err := collection.GetNFT(id)
	require.Nil(t, err)
	require.Equal(t, testNFT.String(), returnedNFT.String())

	returnedNFT, err = collection.GetNFT(id2)
	require.NotNil(t, err)
	require.Nil(t, returnedNFT)
}

func TestCollectionContainsNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	contains := collection.ContainsNFT(id)
	require.True(t, contains)

	contains = collection.ContainsNFT(id2)
	require.False(t, contains)
}

func TestCollectionAddNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	newCollection, err := collection.AddNFT(testNFT)
	require.NotNil(t, err)
	require.Equal(t, collection.String(), newCollection.String())

	newCollection, err = collection.AddNFT(testNFT2)
	require.Nil(t, err)
	require.NotEqual(t, collection.String(), newCollection.String())
	require.Equal(t, len(newCollection.NFTs), 2)

}

func TestCollectionUpdateNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	testNFT3 := NewBaseNFT(id, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	newCollection, err := collection.UpdateNFT(testNFT2)
	require.NotNil(t, err)
	require.Equal(t, collection.String(), newCollection.String())

	collection, err = collection.UpdateNFT(testNFT3)
	require.Nil(t, err)

	var returnedNFT NFT
	returnedNFT, err = collection.GetNFT(id)
	require.Nil(t, err)

	require.Equal(t, returnedNFT.GetOwner(), ownerAddress2)
	require.Equal(t, returnedNFT.GetName(), name2)
	require.Equal(t, returnedNFT.GetDescription(), description2)
	require.Equal(t, returnedNFT.GetImage(), image2)
	require.Equal(t, returnedNFT.GetTokenURI(), tokenURI2)

}

func TestCollectionDeleteNFTMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	testNFT3 := NewBaseNFT(id3, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT, testNFT2)
	collection := NewCollection(denom, nfts)

	newCollection, err := collection.DeleteNFT(testNFT3)
	require.NotNil(t, err)
	require.Equal(t, collection.String(), newCollection.String())

	collection, err = collection.DeleteNFT(testNFT2)
	require.Nil(t, err)
	require.Equal(t, len(collection.NFTs), 1)

	var returnedNFT NFT
	returnedNFT, err = collection.GetNFT(id2)
	require.Nil(t, returnedNFT)
	require.NotNil(t, err)
}

func TestCollectionSupplyMethod(t *testing.T) {

	empty := EmptyCollection()
	require.Equal(t, empty.Supply(), 0)

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts := NewNFTs(testNFT, testNFT2)
	collection := NewCollection(denom, nfts)

	require.Equal(t, collection.Supply(), 2)
}

func TestCollectionStringMethod(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts := NewNFTs(testNFT, testNFT2)
	collection := NewCollection(denom, nfts)
	require.Equal(t, collection.String(),
		fmt.Sprintf(`Denom: 				%s
NFTs:

ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s
ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`, denom, id, ownerAddress.String(), name, description, image, tokenURI,
			id2, ownerAddress2.String(), name2, description2, image2, tokenURI2))
}

// ---------------------------------------- Collections ---------------------------------------------------

func TestNewCollections(t *testing.T) {

	emptyCollections := NewCollections()
	require.Empty(t, emptyCollections)

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts2 := NewNFTs(testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections := NewCollections(collection, collection2)
	require.Equal(t, len(collections), 2)

}

func TestCollectionsAddMethod(t *testing.T) {

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	collections := NewCollections(collection)

	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts2 := NewNFTs(testNFT2)
	collection2 := NewCollection(denom2, nfts2)
	collections2 := NewCollections(collection2)

	collections = collections.Add(collections2)
	require.Equal(t, len(collections), 2)

}
func TestCollectionsFindMethod(t *testing.T) {

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts2 := NewNFTs(testNFT2)
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

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	collections := NewCollections(collection)

	returnedCollections, removed := collections.Remove(denom2)
	require.False(t, removed)
	require.Equal(t, returnedCollections.String(), collections.String())

	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts2 := NewNFTs(testNFT2)
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

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts2 := NewNFTs(testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections = NewCollections(collection, collection2)
	require.Equal(t, fmt.Sprintf(`Denom: 				%s
NFTs:

ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s
Denom: 				%s
NFTs:

ID:				%s
Owner:			%s
Name:			%s
Description: 	%s
Image:			%s
TokenURI:		%s`, denom, id, ownerAddress.String(), name, description, image, tokenURI,
		denom2, id2, ownerAddress2.String(), name2, description2, image2, tokenURI2), collections.String())

}

func TestCollectionsEmptyMethod(t *testing.T) {

	collections := NewCollections()
	require.True(t, collections.Empty())

	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	collections = NewCollections(collection)
	require.False(t, collections.Empty())

}

func TestCollectionsSortInterface(t *testing.T) {
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts2 := NewNFTs(testNFT2)
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
	testNFT := NewBaseNFT(id, ownerAddress, name, description, image, tokenURI)
	nfts := NewNFTs(testNFT)
	collection := NewCollection(denom, nfts)

	testNFT2 := NewBaseNFT(id2, ownerAddress2, name2, description2, image2, tokenURI2)
	nfts2 := NewNFTs(testNFT2)
	collection2 := NewCollection(denom2, nfts2)

	collections := NewCollections(collection, collection2)

	bz, err := collections.MarshalJSON()
	require.Nil(t, err)
	require.Equal(t, string(bz), fmt.Sprintf(`{"%s":{"nfts":{"%s":{"id":"%s","owner":"%s","name":"%s","description":"%s","image":"%s","token_uri":"%s"}}},"%s":{"nfts":{"%s":{"id":"%s","owner":"%s","name":"%s","description":"%s","image":"%s","token_uri":"%s"}}}}`,
		denom, id, id, ownerAddress.String(), name, description, image, tokenURI,
		denom2, id2, id2, ownerAddress2.String(), name2, description2, image2, tokenURI2,
	))

	var newCollections Collections
	err = newCollections.UnmarshalJSON(bz)
	require.Nil(t, err)

	err = newCollections.UnmarshalJSON([]byte{})
	require.NotNil(t, err)
}

// ---------------------------------------- IDCollection ---------------------------------------------------

func TestNewIDCollection(t *testing.T) {
	ids := []string{id, id2, id3}
	idCollection := NewIDCollection(denom, ids)
	require.Equal(t, idCollection.Denom, denom)
	require.Equal(t, len(idCollection.IDs), 3)
}

func TestEmptyIDCollection(t *testing.T) {
	idCollection := EmptyIDCollection()
	require.Empty(t, idCollection.Denom)
	require.Empty(t, idCollection.IDs)
}

func TestIDCollectionExistsMethod(t *testing.T) {
	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	require.True(t, idCollection.Exists(id))
	require.True(t, idCollection.Exists(id2))
	require.False(t, idCollection.Exists(id3))
}

func TestIDCollectionAddIDMethod(t *testing.T) {
	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	idCollection = idCollection.AddID(id3)
	require.Equal(t, len(idCollection.IDs), 3)
}

func TestIDCollectionDeleteIDMethod(t *testing.T) {
	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	newIDCollection, err := idCollection.DeleteID(id3)
	require.NotNil(t, err)
	require.Equal(t, idCollection.String(), newIDCollection.String())

	idCollection, err = idCollection.DeleteID(id2)
	require.Equal(t, len(idCollection.IDs), 1)
}

func TestIDCollectionSupplyMethod(t *testing.T) {
	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	require.Equal(t, 2, idCollection.Supply())

	idCollection = EmptyIDCollection()
	require.Equal(t, 0, idCollection.Supply())
}

func TestIDCollectionStringMethod(t *testing.T) {
	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	require.Equal(t, idCollection.String(), fmt.Sprintf(`Denom: 			%s
IDs:        	%s,%s`, denom, id, id2))
}

// ---------------------------------------- IDCollections ---------------------------------------------------

func TestIDCollectionsString(t *testing.T) {

	emptyCollections := IDCollections([]IDCollection{})
	require.Equal(t, emptyCollections.String(), "")

	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	idCollection2 := NewIDCollection(denom2, ids)

	idCollections := IDCollections([]IDCollection{idCollection, idCollection2})
	require.Equal(t, idCollections.String(), fmt.Sprintf(`Denom: 			%s
IDs:        	%s,%s
Denom: 			%s
IDs:        	%s,%s`, denom, id, id2, denom2, id, id2))
}

// ---------------------------------------- Owner ---------------------------------------------------

func TestNewOwner(t *testing.T) {

	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	idCollection2 := NewIDCollection(denom2, ids)

	owner := NewOwner(ownerAddress, idCollection, idCollection2)
	require.Equal(t, owner.Address.String(), ownerAddress.String())
	require.Equal(t, len(owner.IDCollections), 2)
}

func TestOwnerSupplyMethod(t *testing.T) {

	owner := NewOwner(ownerAddress)
	require.Equal(t, owner.Supply(), 0)

	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	owner = NewOwner(ownerAddress, idCollection)
	require.Equal(t, owner.Supply(), 2)

	idCollection2 := NewIDCollection(denom2, ids)
	owner = NewOwner(ownerAddress, idCollection, idCollection2)
	require.Equal(t, owner.Supply(), 4)
}

func TestOwnerGetIDCollectionMethod(t *testing.T) {

	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	owner := NewOwner(ownerAddress, idCollection)

	gotCollection, found := owner.GetIDCollection(denom2)
	require.False(t, found)
	require.Equal(t, gotCollection.Denom, "")
	require.Equal(t, len(gotCollection.IDs), 0)
	require.Equal(t, gotCollection.String(), EmptyIDCollection().String())

	gotCollection, found = owner.GetIDCollection(denom)
	require.True(t, found)
	require.Equal(t, gotCollection.String(), idCollection.String())

	idCollection2 := NewIDCollection(denom2, ids)
	owner = NewOwner(ownerAddress, idCollection, idCollection2)

	gotCollection, found = owner.GetIDCollection(denom)
	require.True(t, found)
	require.Equal(t, gotCollection.String(), idCollection.String())

	gotCollection, found = owner.GetIDCollection(denom2)
	require.True(t, found)
	require.Equal(t, gotCollection.String(), idCollection2.String())
}

func TestOwnerUpdateIDCollectionMethod(t *testing.T) {
	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	owner := NewOwner(ownerAddress, idCollection)

	ids2 := []string{id, id2, id3}
	idCollection2 := NewIDCollection(denom2, ids2)

	returnedOwner, err := owner.UpdateIDCollection(idCollection2)
	require.NotNil(t, err)
	require.Equal(t, owner.String(), returnedOwner.String())

	idCollection2 = NewIDCollection(denom, ids2)
	returnedOwner, err = owner.UpdateIDCollection(idCollection2)
	require.Nil(t, err)

	returnedCollection, _ := owner.GetIDCollection(denom)
	require.Equal(t, len(returnedCollection.IDs), 3)
}

func TestOwnerDeleteIDMethod(t *testing.T) {
	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	owner := NewOwner(ownerAddress, idCollection)

	returnedOwner, err := owner.DeleteID(denom2, id)
	require.NotNil(t, err)
	require.Equal(t, owner.String(), returnedOwner.String())

	returnedOwner, err = owner.DeleteID(denom, id3)
	require.NotNil(t, err)
	require.Equal(t, owner.String(), returnedOwner.String())

	owner, err = owner.DeleteID(denom, id)
	require.Nil(t, err)

	returnedCollection, _ := owner.GetIDCollection(denom)
	require.Equal(t, len(returnedCollection.IDs), 1)
}
