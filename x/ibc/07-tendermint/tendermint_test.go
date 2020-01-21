package tendermint_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmtypes "github.com/tendermint/tendermint/types"

	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
)

const (
	chainID = "gaia"
	height  = 4
)

type TendermintTestSuite struct {
	suite.Suite

	privVal tmtypes.PrivValidator
	valSet  *tmtypes.ValidatorSet
	header  tendermint.Header
}

func (suite *TendermintTestSuite) SetupTest() {
	suite.privVal = tmtypes.NewMockPV()
	val := tmtypes.NewValidator(suite.privVal.GetPubKey(), 10)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	suite.header = tendermint.CreateTestHeader(chainID, height, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
}

func TestTendermintTestSuite(t *testing.T) {
	suite.Run(t, new(TendermintTestSuite))
}
