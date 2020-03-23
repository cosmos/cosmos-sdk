package types_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/mocks"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestChainAnteDecorators(t *testing.T) {
	t.Parallel()
	// test panic
	require.Nil(t, sdk.ChainAnteDecorators([]sdk.AnteDecorator{}...))

	ctx, tx := sdk.Context{}, sdk.Tx(nil)
	mockCtrl := gomock.NewController(t)
	mockAnteDecorator1 := mocks.NewMockAnteDecorator(mockCtrl)
	mockAnteDecorator1.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, gomock.Any()).Times(1)
	sdk.ChainAnteDecorators(mockAnteDecorator1)(ctx, tx, true)

	mockAnteDecorator2 := mocks.NewMockAnteDecorator(mockCtrl)
	mockAnteDecorator1.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, mockAnteDecorator2).Times(1)
	mockAnteDecorator2.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, nil).Times(1)
	sdk.ChainAnteDecorators(mockAnteDecorator1, mockAnteDecorator2)
}
