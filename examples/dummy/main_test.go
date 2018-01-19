package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/abci/types"
)

func TestQuery(t *testing.T) {
	app, err := newDummyApp()
	require.NoError(t, err)

	k, v := []byte("my-key"), []byte("some-awesome-value")

	// make sure that we reject proof for now
	badReq := abci.RequestQuery{Data: k, Prove: true, Path: "/store"}
	res := app.Query(badReq)
	assert.Equal(t, 500, int(res.Code))

	// check that missing data returns correct code
	req := abci.RequestQuery{Data: k, Path: "/main"}
	res = app.Query(req)
	assert.Equal(t, 404, int(res.Code))

	// submit a tx and commit
	txBytes := append(append(k, '='), v...)
	fmt.Printf("tx: %x\n", txBytes)
	dres := app.DeliverTx(txBytes)
	require.Equal(t, 0, int(dres.Code), dres.Log)
	cres := app.Commit()
	// we want a non-empty hash
	require.NotEqual(t, 0, len(cres.Data))

	// now try to query for existing data
	res = app.Query(req)
	assert.Equal(t, 0, int(res.Code))
	assert.Equal(t, k, res.Key)
	assert.Equal(t, v, res.Value)
}
