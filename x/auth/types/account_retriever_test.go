package types

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/mocks"
)

var dummyError = errors.New("dummy")

func TestAccountRetriever(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockNodeQuerier := mocks.NewMockNodeQuerier(mockCtrl)
	accRetr := NewAccountRetriever(mockNodeQuerier)
	addr := []byte("test")
	bs, err := ModuleCdc.MarshalJSON(NewQueryAccountParams(addr))
	require.NoError(t, err)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq("custom/acc/account"),
		gomock.Eq(bs)).Return(nil, int64(0), dummyError).Times(1)
	accRetr.GetAccount(addr)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq("custom/acc/account"),
		gomock.Eq(bs)).Return(nil, int64(0), dummyError).Times(1)
	accRetr.GetAccountNumberSequence(addr)

	mockNodeQuerier.EXPECT().QueryWithData(gomock.Eq("custom/acc/account"),
		gomock.Eq(bs)).Return(nil, int64(0), dummyError).Times(1)
	accRetr.EnsureExists(addr)
}
