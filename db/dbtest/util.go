package dbtest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	dbm "github.com/cosmos/cosmos-sdk/db"
)

func AssertNext(t *testing.T, itr dbm.Iterator, expected bool) {
	t.Helper()
	require.Equal(t, expected, itr.Next())
}

func AssertDomain(t *testing.T, itr dbm.Iterator, start, end []byte) {
	t.Helper()
	ds, de := itr.Domain()
	assert.Equal(t, start, ds, "checkDomain domain start incorrect")
	assert.Equal(t, end, de, "checkDomain domain end incorrect")
}

func AssertItem(t *testing.T, itr dbm.Iterator, key, value []byte) {
	t.Helper()
	assert.Exactly(t, key, itr.Key())
	assert.Exactly(t, value, itr.Value())
}

func AssertInvalid(t *testing.T, itr dbm.Iterator) {
	t.Helper()
	AssertNext(t, itr, false)
	AssertKeyPanics(t, itr)
	AssertValuePanics(t, itr)
}

func AssertKeyPanics(t *testing.T, itr dbm.Iterator) {
	t.Helper()
	assert.Panics(t, func() { itr.Key() }, "checkKeyPanics expected panic but didn't")
}

func AssertValue(t *testing.T, db dbm.DBReader, key, valueWanted []byte) {
	t.Helper()
	valueGot, err := db.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, valueWanted, valueGot)
}

func AssertValuePanics(t *testing.T, itr dbm.Iterator) {
	t.Helper()
	assert.Panics(t, func() { itr.Value() })
}
