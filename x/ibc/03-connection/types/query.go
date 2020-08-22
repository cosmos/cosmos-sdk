package types

import (
	"strings"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// NewQueryConnectionResponse creates a new QueryConnectionResponse instance
func NewQueryConnectionResponse(
	connectionID string, connection ConnectionEnd, proof []byte, height int64,
) *QueryConnectionResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.ConnectionPath(connectionID), "/"))
	return &QueryConnectionResponse{
		Connection:  &connection,
		Proof:       proof,
		ProofPath:   path.Pretty(),
		ProofHeight: uint64(height),
	}
}

// NewQueryClientConnectionsResponse creates a new ConnectionPaths instance
func NewQueryClientConnectionsResponse(
	clientID string, connectionPaths []string, proof []byte, height int64,
) *QueryClientConnectionsResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.ClientConnectionsPath(clientID), "/"))
	return &QueryClientConnectionsResponse{
		ConnectionPaths: connectionPaths,
		Proof:           proof,
		ProofPath:       path.Pretty(),
		ProofHeight:     uint64(height),
	}
}

// NewQueryClientConnectionsRequest creates a new QueryClientConnectionsRequest instance
func NewQueryClientConnectionsRequest(clientID string) *QueryClientConnectionsRequest {
	return &QueryClientConnectionsRequest{
		ClientId: clientID,
	}
}

// NewQueryConnectionClientStateResponse creates a newQueryConnectionClientStateResponse instance
func NewQueryConnectionClientStateResponse(identifiedClientState clienttypes.IdentifiedClientState, proof []byte, height int64) *QueryConnectionClientStateResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.FullClientPath(identifiedClientState.ClientId, host.ClientStatePath()), "/"))
	return &QueryConnectionClientStateResponse{
		IdentifiedClientState: &identifiedClientState,
		Proof:                 proof,
		ProofPath:             path.Pretty(),
		ProofHeight:           uint64(height),
	}
}

// NewQueryConnectionConsensusStateResponse creates a newQueryConnectionConsensusStateResponse instance
func NewQueryConnectionConsensusStateResponse(clientID string, anyConsensusState *codectypes.Any, consensusStateHeight uint64, proof []byte, height int64) *QueryConnectionConsensusStateResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.FullClientPath(clientID, host.ConsensusStatePath(consensusStateHeight)), "/"))
	return &QueryConnectionConsensusStateResponse{
		ConsensusState: anyConsensusState,
		ClientId:       clientID,
		Proof:          proof,
		ProofPath:      path.Pretty(),
		ProofHeight:    uint64(height),
	}
}
