package types_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/tests/mocks"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type handlerTestSuite struct {
	suite.Suite
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(handlerTestSuite))
}

func (s *handlerTestSuite) SetupSuite() {
	s.T().Parallel()
}

func (s *handlerTestSuite) TestChainAnteDecorators() {
	// test panic
	s.Require().Nil(sdk.ChainAnteDecorators([]sdk.AnteDecorator{}...))

	ctx, tx := sdk.Context{}, sdk.Tx(nil)
	mockCtrl := gomock.NewController(s.T())
	mockAnteDecorator1 := mocks.NewMockAnteDecorator(mockCtrl)
	mockAnteDecorator1.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, gomock.Any()).Times(1)
	_, err := sdk.ChainAnteDecorators(mockAnteDecorator1)(ctx, tx, true)
	s.Require().NoError(err)

	mockAnteDecorator2 := mocks.NewMockAnteDecorator(mockCtrl)
	// NOTE: we can't check that mockAnteDecorator2 is passed as the last argument because
	// ChainAnteDecorators wraps the decorators into closures, so each decorator is
	// receving a closure.
	mockAnteDecorator1.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, gomock.Any()).Times(1)
	mockAnteDecorator2.EXPECT().AnteHandle(gomock.Eq(ctx), gomock.Eq(tx), true, gomock.Any()).Times(1)

	_, err = sdk.ChainAnteDecorators(
		mockAnteDecorator1,
		mockAnteDecorator2)(ctx, tx, true)
	s.Require().NoError(err)
}
