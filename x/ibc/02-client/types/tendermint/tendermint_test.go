package tendermint

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmtypes "github.com/tendermint/tendermint/types"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

type TendermintTestSuite struct {
	suite.Suite

	privVal tmtypes.PrivValidator
	valSet  *tmtypes.ValidatorSet
	header  Header
	cs      ConsensusState
}

func (suite *TendermintTestSuite) SetupTest() {
	privVal := tmtypes.NewMockPV()
	val := tmtypes.NewValidator(privVal.GetPubKey(), 10)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	suite.header = MakeHeader(3, valSet, valSet, []tmtypes.PrivValidator{privVal})
	root := commitment.NewRoot(tmhash.Sum([]byte("my root")))

	cs := ConsensusState{
		ChainID:          "gaia",
		Height:           3,
		Root:             root,
		NextValidatorSet: valSet,
	}

	// set fields in suite
	suite.privVal = privVal
	suite.valSet = valSet
	suite.cs = cs
}

func TestTendermintTestSuite(t *testing.T) {
	suite.Run(t, new(TendermintTestSuite))
}
