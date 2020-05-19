package types_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/mocks"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

var errFoo = errors.New("dummy")

func TestAccountRetriever(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNodeQuerier := mocks.NewMockNodeQuerier(mockCtrl)
	accRetr := types.NewAccountRetriever(appCodec)
	addr := []byte("test")
	bs, err := appCodec.MarshalJSON(types.NewQueryAccountParams(addr))
	require.NoError(t, err)

	route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryAccount)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq(route),
		gomock.Eq(bs)).Return(nil, int64(0), errFoo).Times(1)
	_, err = accRetr.GetAccount(mockNodeQuerier, addr)
	require.Error(t, err)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq(route),
		gomock.Eq(bs)).Return(nil, int64(0), errFoo).Times(1)
	n, s, err := accRetr.GetAccountNumberSequence(mockNodeQuerier, addr)
	require.Error(t, err)
	require.Equal(t, uint64(0), n)
	require.Equal(t, uint64(0), s)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq(route),
		gomock.Eq(bs)).Return(nil, int64(0), errFoo).Times(1)
	require.Error(t, accRetr.EnsureExists(mockNodeQuerier, addr))
}
