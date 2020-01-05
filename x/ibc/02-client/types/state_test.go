package types

import "github.com/stretchr/testify/require"

func (suite *ClientTestSuite) TestNewClientState() {
	state := NewClientState(suite.clientID)
	require.Equal(suite.T(), suite.clientID, state.ID)
	require.Equal(suite.T(), false, state.Frozen)
}