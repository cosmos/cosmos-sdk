package types

import (
	tmtypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/stretchr/testify/require"
)

func (suite *ClientTestSuite) TestNewQueryClientStateParams() {
	q := NewQueryClientStateParams("id")
	require.Equal(suite.T(), "id", q.ClientID)
}

func (suite *ClientTestSuite) TestNewQueryAllClientsParams() {
	q := NewQueryAllClientsParams(4, 5)
	require.Equal(suite.T(), 4, q.Page)
	require.Equal(suite.T(), 5, q.Limit)
}

func (suite *ClientTestSuite) TestNewQueryCommitmentRootParams() {
	q := NewQueryCommitmentRootParams("id", 5)
	require.Equal(suite.T(), "id", q.ClientID)
	require.Equal(suite.T(), uint64(5), q.Height)
}

func (suite *ClientTestSuite) TestNewQueryCommitterParams() {
	q := NewQueryCommitterParams("id", 5)
	require.Equal(suite.T(), "id", q.ClientID)
	require.Equal(suite.T(), uint64(5), q.Height)
}

func (suite *ClientTestSuite) TestNewClientStateResponse() {
	q := NewClientStateResponse("id", State{ID: "state_id", Frozen: false}, nil, 5)
	require.Equal(suite.T(), "state_id", q.ClientState.ID)
	require.Equal(suite.T(), "/1/clients/id/state", q.ProofPath.String())
	require.Equal(suite.T(), uint64(5), q.ProofHeight)
}

func (suite *ClientTestSuite) TestNewConsensusStateResponse() {
	q := NewConsensusStateResponse("id", tmtypes.ConsensusState{ChainID: "chain_id"}, nil, 5)
	require.Equal(suite.T(), "chain_id", q.ConsensusState.ChainID)
	require.Equal(suite.T(), "/3/clients/id/consensusState", q.ProofPath.String())
	require.Equal(suite.T(), uint64(5), q.ProofHeight)
}

func (suite *ClientTestSuite) TestNewRootResponse() {
	q := NewRootResponse("id", 4, commitment.Root{Hash: []byte("hash")}, nil, 5)
	require.Equal(suite.T(), "hash", string(q.Root.Hash))
	require.Equal(suite.T(), "/4/clients/id/roots/4", q.ProofPath.String())
	require.Equal(suite.T(), uint64(5), q.ProofHeight)
}

func (suite *ClientTestSuite) TestNewCommitterResponse() {
	q := NewCommitterResponse("id", 4, tmtypes.Committer{NextValSetHash: []byte("hash")}, nil, 5)
	require.Equal(suite.T(), "hash", string(q.Committer.NextValSetHash))
	require.Equal(suite.T(), "/5/clients/id/committer/4", q.ProofPath.String())
	require.Equal(suite.T(), uint64(5), q.ProofHeight)
}
