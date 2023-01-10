package types_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestChainAnteDecorators(t *testing.T) {
	// test panic
	require.Nil(t, sdk.ChainAnteDecorators([]sdk.AnteDecorator{}...))

	ctx, tx := sdk.Context{}, sdk.Tx(nil)
	mockCtrl := gomock.NewController(t)
	mockAnteDecorator1 := mock.NewMockAnteDecorator(mockCtrl)
	mockAnteDecorator1.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, gomock.Any()).Times(1)
	_, err := sdk.ChainAnteDecorators(mockAnteDecorator1)(ctx, tx, true)
	require.NoError(t, err)

	mockAnteDecorator2 := mock.NewMockAnteDecorator(mockCtrl)
	// NOTE: we can't check that mockAnteDecorator2 is passed as the last argument because
	// ChainAnteDecorators wraps the decorators into closures, so each decorator is
	// receiving a closure.
	mockAnteDecorator1.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, gomock.Any()).Times(1)
	mockAnteDecorator2.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, gomock.Any()).Times(1)

	_, err = sdk.ChainAnteDecorators(
		mockAnteDecorator1,
		mockAnteDecorator2)(ctx, tx, true)
	require.NoError(t, err)
}
