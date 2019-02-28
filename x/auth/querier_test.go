package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func Test_queryAccount(t *testing.T) {
	input := setupTestInput()
	req := abci.RequestQuery{
		Path: fmt.Sprintf("custom/%s/%s", QuerierRoute, QueryAccount),
		Data: []byte{},
	}

	res, err := queryAccount(input.ctx, req, input.ak)
	require.NotNil(t, err)
	require.Nil(t, res)

	req.Data = input.cdc.MustMarshalJSON(NewQueryAccountParams([]byte("")))
	res, err = queryAccount(input.ctx, req, input.ak)
	require.Nil(t, err)
	require.Equal(t, res, []byte("null"))

	_, _, addr := keyPubAddr()
	req.Data = input.cdc.MustMarshalJSON(NewQueryAccountParams(addr))
	res, err = queryAccount(input.ctx, req, input.ak)
	require.Nil(t, err)
	require.Equal(t, res, []byte("null"))

	input.ak.SetAccount(input.ctx, input.ak.NewAccountWithAddress(input.ctx, addr))
	res, err = queryAccount(input.ctx, req, input.ak)
	require.Nil(t, err)
	require.NotEqual(t, res, []byte("null"))

	var account Account
	err2 := input.cdc.UnmarshalJSON(res, &account)
	require.Nil(t, err2)
}
