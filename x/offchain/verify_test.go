package offchain

import (
	"errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type verifyMessageTestSuite struct {
	suite.Suite
	address   sdk.AccAddress
	validData []byte
}

func (ts *verifyMessageTestSuite) TestValidMessage() {
	err := VerifyMessage(NewMsgSignData(ts.address, ts.validData))
	ts.Require().NoError(err, "message should be valid")
}

func (ts *verifyMessageTestSuite) TestInvalidMessageType() {
	err := VerifyMessage(&types.MsgSend{})
	ts.Require().True(errors.Is(err, errInvalidRoute), "unexpected error: %s", err)
}

func (ts *verifyMessageTestSuite) TestInvalidRoute() {
	err := VerifyMessage()
	ts.Require().True(errors.Is(err, errInvalidRoute), "unexpected error: %s", err)
}

func TestVerifyMessage(t *testing.T) {
	suite.Run(t, new(verifyMessageTestSuite))
}
