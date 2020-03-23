package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	chainID                      = "gaia"
	height                       = 4
	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
)

type TendermintTestSuite struct {
	suite.Suite

	cdc     *codec.Codec
	privVal tmtypes.PrivValidator
	valSet  *tmtypes.ValidatorSet
	header  ibctmtypes.Header
	now     time.Time
}

func (suite *TendermintTestSuite) SetupTest() {
	suite.cdc = codec.New()
	codec.RegisterCrypto(suite.cdc)
	ibctmtypes.RegisterCodec(suite.cdc)
	commitmenttypes.RegisterCodec(suite.cdc)

	suite.now = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.privVal = tmtypes.NewMockPV()
	val := tmtypes.NewValidator(suite.privVal.GetPubKey(), 10)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	suite.header = ibctmtypes.CreateTestHeader(chainID, height, suite.now, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
}

func TestTendermintTestSuite(t *testing.T) {
	suite.Run(t, new(TendermintTestSuite))
}
