package types_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/07-tendermint/types"
)

func (suite *TendermintTestSuite) TestExportMetadata() {
	clientState := types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs(), upgradePath, false, false)
	suite.chainA.App.IBCKeeper.ClientKeeper.SetClientState(suite.chainA.GetContext(), "clientA", clientState)

	gm := clientState.ExportMetadata(suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), "clientA"))
	suite.Require().Nil(gm, "client with no metadata returned non-nil exported metadata")

	clientStore := suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), "clientA")

	// set some processed times
	timestamp1 := uint64(time.Now().UnixNano())
	timestamp2 := uint64(time.Now().Add(time.Minute).UnixNano())
	timestampBz1 := sdk.Uint64ToBigEndian(timestamp1)
	timestampBz2 := sdk.Uint64ToBigEndian(timestamp2)
	types.SetProcessedTime(clientStore, clienttypes.NewHeight(0, 1), timestamp1)
	types.SetProcessedTime(clientStore, clienttypes.NewHeight(0, 2), timestamp2)

	gm = clientState.ExportMetadata(suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), "clientA"))
	suite.Require().NotNil(gm, "client with metadata returned nil exported metadata")
	suite.Require().Len(gm, 2, "exported metadata has unexpected length")

	suite.Require().Equal(types.ProcessedTimeKey(clienttypes.NewHeight(0, 1)), gm[0].GetKey(), "metadata has unexpected key")
	suite.Require().Equal(timestampBz1, gm[0].GetValue(), "metadata has unexpected value")

	suite.Require().Equal(types.ProcessedTimeKey(clienttypes.NewHeight(0, 2)), gm[1].GetKey(), "metadata has unexpected key")
	suite.Require().Equal(timestampBz2, gm[1].GetValue(), "metadata has unexpected value")
}
