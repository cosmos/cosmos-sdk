package types

import (
	"strings"

	"github.com/tendermint/tendermint/crypto/merkle"

	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// query routes supported by the IBC connection Querier
const (
	QueryAllConnections       = "connections"
	QueryClientConnections    = "client_connections"
	QueryAllClientConnections = "all_client_connections"
)

// NewQueryConnectionResponse creates a new QueryConnectionResponse instance
func NewQueryConnectionResponse(
	connectionID string, connection ConnectionEnd, proof *merkle.Proof, height int64,
) *QueryConnectionResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.ConnectionPath(connectionID), "/"))
	return &QueryConnectionResponse{
		Connection:  &connection,
		Proof:       &commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:   &path,
		ProofHeight: uint64(height),
	}
}

// QueryAllConnectionsParams defines the parameters necessary for querying for all
// connections.
// Deprecated.
type QueryAllConnectionsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryAllConnectionsParams creates a new QueryAllConnectionsParams instance.
// Deprecated:
func NewQueryAllConnectionsParams(page, limit int) QueryAllConnectionsParams {
	return QueryAllConnectionsParams{
		Page:  page,
		Limit: limit,
	}
}

// NewQueryClientConnectionsResponse creates a new ConnectionPaths instance
func NewQueryClientConnectionsResponse(
	clientID string, connectionPaths []string, proof *merkle.Proof, height int64,
) *QueryClientConnectionsResponse {
	path := commitmenttypes.NewMerklePath(strings.Split(host.ClientConnectionsPath(clientID), "/"))
	return &QueryClientConnectionsResponse{
		ConnectionPaths: connectionPaths,
		Proof:           &commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:       &path,
		ProofHeight:     uint64(height),
	}
}

// NewQueryClientConnectionsRequest creates a new QueryClientConnectionsRequest instance
func NewQueryClientConnectionsRequest(clientID string) QueryClientConnectionsRequest {
	return QueryClientConnectionsRequest{
		ClientID: clientID,
	}
}
