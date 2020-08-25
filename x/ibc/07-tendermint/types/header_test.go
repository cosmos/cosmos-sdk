package types_test

import (
	"time"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

func (suite *TendermintTestSuite) TestGetHeight() {
	header := suite.chainA.LastHeader
	suite.Require().NotEqual(uint64(0), header.GetHeight())

	header.Header = nil
	suite.Require().Equal(uint64(0), header.GetHeight())
}

func (suite *TendermintTestSuite) TestGetTime() {
	header := suite.chainA.LastHeader
	suite.Require().NotEqual(time.Time{}, header.GetTime())

	header.Header = nil
	suite.Require().Equal(time.Time{}, header.GetTime())
}

func (suite *TendermintTestSuite) TestHeaderValidateBasic() {
	var (
		header  *types.Header
		chainID string
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"valid header", func() {}, true},
		{"signed header is nil", func() {
			header.Header = nil
		}, false},
		{"signed header failed tendermint ValidateBasic", func() {
			header = suite.chainA.LastHeader
			chainID = "chainid"
		}, false},
		{"trusted height is greater than header height", func() {
			header.TrustedHeight = header.GetHeight() + 1
		}, false},
		{"validator set nil", func() {
			header.ValidatorSet = nil
		}, false},
		{"header validator hash does not equal hash of validator set", func() {
			// use chainB's randomly generated validator set
			header.ValidatorSet = suite.chainB.LastHeader.ValidatorSet
		}, false},
	}

	suite.Require().Equal(clientexported.Tendermint, suite.header.ClientType())

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			chainID = suite.chainA.ChainID   // must be explicitly changed in malleate
			header = suite.chainA.LastHeader // must be explicitly changed in malleate

			tc.malleate()

			err := header.ValidateBasic(chainID)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
