package types

import (
	"github.com/stretchr/testify/require"

	ibctypes "github.com/cosmos/cosmos-sdk/x/ibc/types"
)

func (suite *ClientTestSuite) TestClientStatePath() {
	expected := "1/clients/tendermint/state"
	require.Equal(suite.T(), expected, ClientStatePath(suite.clientID))
}

func (suite *ClientTestSuite) TestClientTypePath() {
	expected := "2/clients/tendermint/type"
	require.Equal(suite.T(), expected, ClientTypePath(suite.clientID))
}

func (suite *ClientTestSuite) TestConsensusStatePath() {
	expected := "3/clients/tendermint/consensusState"
	require.Equal(suite.T(), expected, ConsensusStatePath(suite.clientID))
}

func (suite *ClientTestSuite) TestRootPath() {
	expected := "4/clients/tendermint/roots/42"
	require.Equal(suite.T(), expected, RootPath(suite.clientID, 42))
}

func (suite *ClientTestSuite) TestCommitterPath() {
	expected := "5/clients/tendermint/committer/42"
	require.Equal(suite.T(), expected, CommitterPath(suite.clientID, 42))
}

func (suite *ClientTestSuite) TestGetClientKeysPrefix() {
	expected := []byte("4/clients")
	require.Equal(suite.T(), expected, GetClientKeysPrefix(ibctypes.KeyRootPrefix))
}