package types_test

import (
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"

	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
)

func (suite *SoloMachineTestSuite) TestEvidence() {
	ev := suite.Evidence()

	suite.Require().Equal(clientexported.SoloMachine, ev.ClientType())
	suite.Require().Equal(suite.clientID, ev.GetClientID())
	suite.Require().Equal("client", ev.Route())
	suite.Require().Equal("client_misbehaviour", ev.Type())
	suite.Require().Equal(tmbytes.HexBytes(tmhash.Sum(solomachinetypes.SubModuleCdc.MustMarshalBinaryBare(ev))), ev.Hash())
	suite.Require().Equal(int64(suite.sequence), ev.GetHeight())
}
