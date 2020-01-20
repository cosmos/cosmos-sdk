package tendermint

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmtypes "github.com/tendermint/tendermint/types"
)

type TendermintTestSuite struct {
	suite.Suite

	privVal tmtypes.PrivValidator
	valSet  *tmtypes.ValidatorSet
	header  Header
}

func (suite *TendermintTestSuite) SetupTest() {
	suite.privVal = tmtypes.NewMockPV()
	val := tmtypes.NewValidator(suite.privVal.GetPubKey(), 10)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	suite.header = CreateTestHeader("gaia", 4, suite.valSet, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
}

func TestTendermintTestSuite(t *testing.T) {
	suite.Run(t, new(TendermintTestSuite))
}
