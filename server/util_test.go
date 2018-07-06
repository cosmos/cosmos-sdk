package server

import (
	"encoding/json"
	"testing"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/stretchr/testify/require"
)

func TestInsertKeyJSON(t *testing.T) {
	cdc := wire.NewCodec()

	foo := map[string]string{"foo": "foofoo"}
	bar := map[string]string{"barInner": "barbar"}

	// create raw messages
	bz, err := cdc.MarshalJSON(foo)
	require.NoError(t, err)
	fooRaw := json.RawMessage(bz)

	bz, err = cdc.MarshalJSON(bar)
	require.NoError(t, err)
	barRaw := json.RawMessage(bz)

	// make the append
	appBz, err := InsertKeyJSON(cdc, fooRaw, "barOuter", barRaw)
	require.NoError(t, err)

	// test the append
	var appended map[string]json.RawMessage
	err = cdc.UnmarshalJSON(appBz, &appended)
	require.NoError(t, err)

	var resBar map[string]string
	err = cdc.UnmarshalJSON(appended["barOuter"], &resBar)
	require.NoError(t, err)

	require.Equal(t, bar, resBar, "appended: %v", appended)
}
