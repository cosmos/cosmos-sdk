package types_test

import (
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

func (suite *TendermintTestSuite) TestHeaderValidateBasic() {
	var (
		header  types.Header
		chainID string
	)
	testCases := []struct {
		name     string
		malleate func()
		expPass  bool
	}{
		{"valid header", func() {
			//		suite.coordinator.SetupClients(suite.chainA, suite.chainB, clientexported.Tendermint)
			header = suite.chainA.LastHeader
		}, true},
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
			header.Header.ValidatorsHash = []byte("validator set")
		}, false},
	}

	suite.Require().Equal(clientexported.Tendermint, suite.header.ClientType())

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			chainID = suite.chainA.ChainID // must be explicitly changed in malleate

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
