package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// ---------------------------------------- IDCollection ---------------------------------------------------

func TestNewIDCollection(t *testing.T) {
	ids := []string{id, id2, id3}
	idCollection := NewIDCollection(denom, ids)
	require.Equal(t, idCollection.Denom, denom)
	require.Equal(t, len(idCollection.IDs), 3)
}

func TestIDCollectionExistsMethod(t *testing.T) {
	ids := []string{id2, id}
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
	require.Error(t, err)
	require.Equal(t, idCollection.String(), newIDCollection.String())

	idCollection, err = idCollection.DeleteID(id2)
	require.NoError(t, err)
	require.Equal(t, len(idCollection.IDs), 1)
}

func TestIDCollectionSupplyMethod(t *testing.T) {
	idCollectionEmpty := IDCollection{}
	require.Equal(t, 0, idCollectionEmpty.Supply())

	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	require.Equal(t, 2, idCollection.Supply())

	idCollection, err := idCollection.DeleteID(id)
	require.Nil(t, err)
	require.Equal(t, idCollection.Supply(), 1)

	idCollection, err = idCollection.DeleteID(id2)
	require.Nil(t, err)
	require.Equal(t, idCollection.Supply(), 0)

	idCollection = idCollection.AddID(id)
	require.Nil(t, err)
	require.Equal(t, idCollection.Supply(), 1)

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

	owner := NewOwner(address, idCollection, idCollection2)
	require.Equal(t, owner.Address.String(), address.String())
	require.Equal(t, len(owner.IDCollections), 2)
}

func TestOwnerSupplyMethod(t *testing.T) {

	owner := NewOwner(address)
	require.Equal(t, owner.Supply(), 0)

	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	owner = NewOwner(address, idCollection)
	require.Equal(t, owner.Supply(), 2)

	idCollection2 := NewIDCollection(denom2, ids)
	owner = NewOwner(address, idCollection, idCollection2)
	require.Equal(t, owner.Supply(), 4)
}

func TestOwnerGetIDCollectionMethod(t *testing.T) {

	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	owner := NewOwner(address, idCollection)

	gotCollection, found := owner.GetIDCollection(denom2)
	require.False(t, found)
	require.Equal(t, gotCollection.Denom, "")
	require.Equal(t, len(gotCollection.IDs), 0)
	require.Equal(t, gotCollection.String(), IDCollection{}.String())

	gotCollection, found = owner.GetIDCollection(denom)
	require.True(t, found)
	require.Equal(t, gotCollection.String(), idCollection.String())

	idCollection2 := NewIDCollection(denom2, ids)
	owner = NewOwner(address, idCollection, idCollection2)

	gotCollection, found = owner.GetIDCollection(denom)
	require.True(t, found)
	require.Equal(t, gotCollection.String(), idCollection.String())

	gotCollection, found = owner.GetIDCollection(denom2)
	require.True(t, found)
	require.Equal(t, gotCollection.String(), idCollection2.String())
}

func TestOwnerUpdateIDCollectionMethod(t *testing.T) {
	ids := []string{id}
	idCollection := NewIDCollection(denom, ids)
	owner := NewOwner(address, idCollection)
	require.Equal(t, owner.Supply(), 1)

	ids2 := []string{id, id2}
	idCollection2 := NewIDCollection(denom2, ids2)

	// UpdateIDCollection should fail if denom doesn't exist
	returnedOwner, err := owner.UpdateIDCollection(idCollection2)
	require.Error(t, err)

	idCollection3 := NewIDCollection(denom, ids2)
	returnedOwner, err = owner.UpdateIDCollection(idCollection3)
	require.NoError(t, err)
	require.Equal(t, returnedOwner.Supply(), 2)

	owner = returnedOwner

	returnedCollection, _ := owner.GetIDCollection(denom)
	require.Equal(t, len(returnedCollection.IDs), 2)

	owner = NewOwner(address, idCollection, idCollection2)
	require.Equal(t, owner.Supply(), 3)

	returnedOwner, err = owner.UpdateIDCollection(idCollection3)
	require.NoError(t, err)
	require.Equal(t, returnedOwner.Supply(), 4)

}

func TestOwnerDeleteIDMethod(t *testing.T) {
	ids := []string{id, id2}
	idCollection := NewIDCollection(denom, ids)
	owner := NewOwner(address, idCollection)

	returnedOwner, err := owner.DeleteID(denom2, id)
	require.Error(t, err)
	require.Equal(t, owner.String(), returnedOwner.String())

	returnedOwner, err = owner.DeleteID(denom, id3)
	require.Error(t, err)
	require.Equal(t, owner.String(), returnedOwner.String())

	owner, err = owner.DeleteID(denom, id)
	require.NoError(t, err)

	returnedCollection, _ := owner.GetIDCollection(denom)
	require.Equal(t, len(returnedCollection.IDs), 1)
}
