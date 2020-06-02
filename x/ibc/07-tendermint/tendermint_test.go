package tendermint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

const (
	chainID                      = "gaia"
	height                       = 4
	trustingPeriod time.Duration = time.Hour * 24 * 7 * 2
	ubdPeriod      time.Duration = time.Hour * 24 * 7 * 3
	maxClockDrift  time.Duration = time.Second * 10
)

type TendermintTestSuite struct {
	suite.Suite

	cdc        *codec.Codec
	privVal    tmtypes.PrivValidator
	valSet     *tmtypes.ValidatorSet
	header     ibctmtypes.Header
	now        time.Time
	clientTime time.Time
	headerTime time.Time
}

func (suite *TendermintTestSuite) SetupTest() {
	suite.cdc = codec.New()
	cryptocodec.RegisterCrypto(suite.cdc)
	ibctmtypes.RegisterCodec(suite.cdc)
	commitmenttypes.RegisterCodec(suite.cdc)

	// now is the time of the current chain, must be after the updating header
	// mocks ctx.BlockTime()
	suite.now = time.Date(2020, 1, 3, 0, 0, 0, 0, time.UTC)
	suite.clientTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	// Header time is intended to be time for any new header used for updates
	suite.headerTime = time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)
	suite.privVal = tmtypes.NewMockPV()

	pubKey, err := suite.privVal.GetPubKey()
	suite.Require().NoError(err)

	val := tmtypes.NewValidator(pubKey, 10)
	suite.valSet = tmtypes.NewValidatorSet([]*tmtypes.Validator{val})
	// Suite header is intended to be header passed in for initial ClientState
	// Thus it should have same height and time as ClientState
	suite.header = ibctmtypes.CreateTestHeader(chainID, height, suite.clientTime, suite.valSet, []tmtypes.PrivValidator{suite.privVal})
}

func TestTendermintTestSuite(t *testing.T) {
	suite.Run(t, new(TendermintTestSuite))
}
