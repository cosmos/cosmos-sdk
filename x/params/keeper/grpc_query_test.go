package keeper_test

import (
	gocontext "context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/require"
)

func TestGRPCQueryParams(t *testing.T) {
	_, ctx, _, _, keeper := testComponents()

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	proposal.RegisterQueryServer(queryHelper, keeper)
	queryClient := proposal.NewQueryClient(queryHelper)

	res, err := queryClient.Parameters(gocontext.Background(), &proposal.QueryParametersRequest{})
	require.Error(t, err)
	require.Nil(t, res)

	res, err = queryClient.Parameters(gocontext.Background(), &proposal.QueryParametersRequest{Subspace: "test"})
	require.Error(t, err)
	require.Nil(t, res)

	res, err = queryClient.Parameters(gocontext.Background(), &proposal.QueryParametersRequest{Subspace: "test", Key: "key"})
	require.Error(t, err)
	require.Nil(t, res)

	key := []byte("key")
	space := keeper.Subspace("test").WithKeyTable(types.NewKeyTable(types.NewParamSetPair(key, paramJSON{}, validateNoOp)))

	res, err = queryClient.Parameters(gocontext.Background(), &proposal.QueryParametersRequest{Subspace: "test", Key: "key"})
	require.NoError(t, err)
	require.Equal(t, res.Params.Value, "")

	err = space.Update(ctx, key, []byte(`{"param1":"10241024"}`))
	require.NoError(t, err)

	res, err = queryClient.Parameters(gocontext.Background(), &proposal.QueryParametersRequest{Subspace: "test", Key: "key"})
	require.NoError(t, err)
	require.Equal(t, string(res.Params.Value), `{"param1":"10241024"}`)

	res, err = queryClient.Parameters(gocontext.Background(), &proposal.QueryParametersRequest{Subspace: "test1", Key: "key"})
	require.Error(t, err)
	require.Nil(t, res)
}
