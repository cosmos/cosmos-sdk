package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
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

	req.Data = input.cdc.MustMarshalJSON(sdk.NewQueryAccountParams([]byte("")))
	res, err = queryAccount(input.ctx, req, input.ak)
	require.NotNil(t, err)
	require.Nil(t, res)

	_, _, addr := types.KeyTestPubAddr()
	req.Data = input.cdc.MustMarshalJSON(sdk.NewQueryAccountParams(addr))
	res, err = queryAccount(input.ctx, req, input.ak)
	require.NotNil(t, err)
	require.Nil(t, res)

	input.ak.SetAccount(input.ctx, input.ak.NewAccountWithAddress(input.ctx, addr))
	res, err = queryAccount(input.ctx, req, input.ak)
	require.Nil(t, err)
	require.NotNil(t, res)

	var account Account
	err2 := input.cdc.UnmarshalJSON(res, &account)
	require.Nil(t, err2)
}
