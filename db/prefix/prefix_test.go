package prefix_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/db"
	"github.com/cosmos/cosmos-sdk/db/dbtest"
	"github.com/cosmos/cosmos-sdk/db/memdb"
	pfx "github.com/cosmos/cosmos-sdk/db/prefix"
)

func fillDBWithStuff(t *testing.T, dbw db.DBWriter) {
	// Under "key" prefix
	require.NoError(t, dbw.Set([]byte("key"), []byte("value")))
	require.NoError(t, dbw.Set([]byte("key1"), []byte("value1")))
	require.NoError(t, dbw.Set([]byte("key2"), []byte("value2")))
	require.NoError(t, dbw.Set([]byte("key3"), []byte("value3")))
	require.NoError(t, dbw.Set([]byte("something"), []byte("else")))
	require.NoError(t, dbw.Set([]byte("k"), []byte("val")))
	require.NoError(t, dbw.Set([]byte("ke"), []byte("valu")))
	require.NoError(t, dbw.Set([]byte("kee"), []byte("valuu")))
	require.NoError(t, dbw.Commit())
}

func mockDBWithStuff(t *testing.T) db.DBConnection {
	dbm := memdb.NewDB()
	fillDBWithStuff(t, dbm.Writer())
	return dbm
}

func makePrefixReader(t *testing.T, dbc db.DBConnection, pre []byte) db.DBReader {
	view := dbc.Reader()
	require.NotNil(t, view)
	return pfx.NewPrefixReader(view, pre)
}

func TestPrefixDBSimple(t *testing.T) {
	pdb := makePrefixReader(t, mockDBWithStuff(t), []byte("key"))

	dbtest.AssertValue(t, pdb, []byte("key"), nil)
	dbtest.AssertValue(t, pdb, []byte("key1"), nil)
	dbtest.AssertValue(t, pdb, []byte("1"), []byte("value1"))
	dbtest.AssertValue(t, pdb, []byte("key2"), nil)
	dbtest.AssertValue(t, pdb, []byte("2"), []byte("value2"))
	dbtest.AssertValue(t, pdb, []byte("key3"), nil)
	dbtest.AssertValue(t, pdb, []byte("3"), []byte("value3"))
	dbtest.AssertValue(t, pdb, []byte("something"), nil)
	dbtest.AssertValue(t, pdb, []byte("k"), nil)
	dbtest.AssertValue(t, pdb, []byte("ke"), nil)
	dbtest.AssertValue(t, pdb, []byte("kee"), nil)
}

func TestPrefixDBIterator1(t *testing.T) {
	pdb := makePrefixReader(t, mockDBWithStuff(t), []byte("key"))

	itr, err := pdb.Iterator(nil, nil)
	require.NoError(t, err)
	dbtest.AssertDomain(t, itr, nil, nil)
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("1"), []byte("value1"))
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("2"), []byte("value2"))
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("3"), []byte("value3"))
	dbtest.AssertNext(t, itr, false)
	dbtest.AssertInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator1(t *testing.T) {
	pdb := makePrefixReader(t, mockDBWithStuff(t), []byte("key"))

	itr, err := pdb.ReverseIterator(nil, nil)
	require.NoError(t, err)
	dbtest.AssertDomain(t, itr, nil, nil)
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("3"), []byte("value3"))
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("2"), []byte("value2"))
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("1"), []byte("value1"))
	dbtest.AssertNext(t, itr, false)
	dbtest.AssertInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator5(t *testing.T) {
	pdb := makePrefixReader(t, mockDBWithStuff(t), []byte("key"))

	itr, err := pdb.ReverseIterator([]byte("1"), nil)
	require.NoError(t, err)
	dbtest.AssertDomain(t, itr, []byte("1"), nil)
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("3"), []byte("value3"))
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("2"), []byte("value2"))
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("1"), []byte("value1"))
	dbtest.AssertNext(t, itr, false)
	dbtest.AssertInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator6(t *testing.T) {
	pdb := makePrefixReader(t, mockDBWithStuff(t), []byte("key"))

	itr, err := pdb.ReverseIterator([]byte("2"), nil)
	require.NoError(t, err)
	dbtest.AssertDomain(t, itr, []byte("2"), nil)
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("3"), []byte("value3"))
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("2"), []byte("value2"))
	dbtest.AssertNext(t, itr, false)
	dbtest.AssertInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBReverseIterator7(t *testing.T) {
	pdb := makePrefixReader(t, mockDBWithStuff(t), []byte("key"))

	itr, err := pdb.ReverseIterator(nil, []byte("2"))
	require.NoError(t, err)
	dbtest.AssertDomain(t, itr, nil, []byte("2"))
	dbtest.AssertNext(t, itr, true)
	dbtest.AssertItem(t, itr, []byte("1"), []byte("value1"))
	dbtest.AssertNext(t, itr, false)
	dbtest.AssertInvalid(t, itr)
	itr.Close()
}

func TestPrefixDBViewVersion(t *testing.T) {
	prefix := []byte("key")
	dbm := memdb.NewDB()
	fillDBWithStuff(t, dbm.Writer())
	id, err := dbm.SaveNextVersion()
	require.NoError(t, err)
	pdb := pfx.NewPrefixReadWriter(dbm.ReadWriter(), prefix)

	pdb.Set([]byte("1"), []byte("newvalue1"))
	pdb.Delete([]byte("2"))
	pdb.Set([]byte("4"), []byte("newvalue4"))
	pdb.Discard()

	dbview, err := dbm.ReaderAt(id)
	require.NotNil(t, dbview)
	require.NoError(t, err)
	view := pfx.NewPrefixReader(dbview, prefix)
	require.NotNil(t, view)
	defer view.Discard()

	dbtest.AssertValue(t, view, []byte("1"), []byte("value1"))
	dbtest.AssertValue(t, view, []byte("2"), []byte("value2"))
	dbtest.AssertValue(t, view, []byte("4"), nil)
}
