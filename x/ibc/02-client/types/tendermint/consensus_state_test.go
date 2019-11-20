package tendermint

import (
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (suite *TendermintTestSuite) TestCheckValidity() {
	// valid header
	err := suite.cs.checkValidity(suite.header)
	require.Nil(suite.T(), err, "validity failed")

	// switch out header ValidatorsHash
	suite.header.ValidatorsHash = tmhash.Sum([]byte("hello"))
	err = suite.cs.checkValidity(suite.header)
	require.NotNil(suite.T(), err, "validator hash is wrong")

	// reset suite and make header.NextValidatorSet different
	// from NextValidatorSetHash
	suite.SetupTest()
	privVal := tmtypes.NewMockPV()
	val := tmtypes.NewValidator(privVal.GetPubKey(), 5)
	suite.header.NextValidatorSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	err = suite.cs.checkValidity(suite.header)
	require.NotNil(suite.T(), err, "header's next validator set is not consistent with hash")

	// reset and make header fail validatebasic
	suite.SetupTest()
	suite.header.ChainID = "not_gaia"
	err = suite.cs.checkValidity(suite.header)
	require.NotNil(suite.T(), err, "invalid header should fail ValidateBasic")
}

func (suite *TendermintTestSuite) TestCheckUpdate() {
	// valid header should successfully update consensus state
	cs, err := suite.cs.CheckValidityAndUpdateState(suite.header)

	require.Nil(suite.T(), err, "valid update failed")
	require.Equal(suite.T(), suite.header.GetHeight(), cs.GetHeight(), "height not updated")
	require.Equal(suite.T(), suite.header.AppHash.Bytes(), cs.GetRoot().GetHash(), "root not updated")
	tmCS, _ := cs.(ConsensusState)
	require.Equal(suite.T(), suite.header.NextValidatorSet, tmCS.NextValidatorSet, "validator set did not update")

	// make header invalid so update should be unsuccessful
	suite.SetupTest()
	suite.header.ChainID = "not_gaia"

	cs, err = suite.cs.CheckValidityAndUpdateState(suite.header)
	require.NotNil(suite.T(), err)
	require.Nil(suite.T(), cs)
}
