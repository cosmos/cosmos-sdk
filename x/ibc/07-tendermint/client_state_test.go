package tendermint_test

import (
	connection "github.com/cosmos/cosmos-sdk/x/ibc/03-connection"
	tendermint "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint"
	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func (suite *TendermintTestSuite) TestVerifyClientConsensusState() {
	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		consensusState tendermint.ConsensusState
		height         uint64
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:        "successful verification",
		// 	clientState: tendermint.NewClientState(chainID, height),
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyClientConsensusState(
			suite.cdc, tc.height, tc.prefix, tc.proof, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}

func (suite *TendermintTestSuite) TestVerifyConnectionState() {
	testCases := []struct {
		name           string
		clientState    tendermint.ClientState
		connectionID   string
		connection     connection.ConnectionEnd
		consensusState tendermint.ConsensusState
		height         uint64
		prefix         commitment.Prefix
		proof          commitment.Proof
		expPass        bool
	}{
		// {
		// 	name:        "successful verification",
		// 	clientState: tendermint.NewClientState(chainID, height),
		// 	consensusState: tendermint.ConsensusState{
		// 		Root: commitment.NewRoot(suite.header.AppHash),
		// 	},
		// 	height:  height,
		// 	prefix:  commitment.NewPrefix([]byte("ibc")),
		// 	expPass: true,
		// },
		{
			name:        "ApplyPrefix failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.Prefix{},
			expPass: false,
		},
		{
			name:        "latest client height < height",
			clientState: tendermint.NewClientState(chainID, height-1),
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "client is frozen",
			clientState: tendermint.ClientState{ID: chainID, LatestHeight: height, FrozenHeight: height - 1},
			consensusState: tendermint.ConsensusState{
				Root: commitment.NewRoot(suite.header.AppHash),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			expPass: false,
		},
		{
			name:        "proof verification failed",
			clientState: tendermint.NewClientState(chainID, height),
			consensusState: tendermint.ConsensusState{
				Root:             commitment.NewRoot(suite.header.AppHash),
				ValidatorSetHash: suite.valSet.Hash(),
			},
			height:  height,
			prefix:  commitment.NewPrefix([]byte("ibc")),
			proof:   commitment.Proof{},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc

		err := tc.clientState.VerifyConnectionState(
			suite.cdc, tc.height, tc.prefix, tc.proof, tc.connectionID, tc.connection, tc.consensusState,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
