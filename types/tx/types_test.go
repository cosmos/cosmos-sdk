package tx_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	"github.com/stretchr/testify/suite"
)

type testMsgSuite struct {
	suite.Suite
}

func TestValidateBasicMessages(t *testing.T) {
	suite.Run(t, new(testMsgSuite))
}

func (s *testMsgSuite) TestMsg() {
	cases := []struct {
		signer []byte
		expErr bool
	}{
		{
			[]byte(""),
			true,
		},
		{
			[]byte("validAddress"),
			false,
		},
	}

	for _, c := range cases {
		msg := testdata.NewTestMsg(c.signer)
		err := msg.ValidateBasic()
		if c.expErr {
			s.Require().Error(err)
		} else {
			s.Require().NoError(err)
		}
	}
}
