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

// ConnectionResponse defines the client query response for a connection which
// also includes a proof and the height from which the proof was retrieved.
type ConnectionResponse struct {
	Connection  ConnectionEnd               `json:"connection" yaml:"connection"`
	Proof       commitmenttypes.MerkleProof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitmenttypes.MerklePath  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64                      `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewConnectionResponse creates a new ConnectionResponse instance
func NewConnectionResponse(
	connectionID string, connection ConnectionEnd, proof *merkle.Proof, height int64,
) ConnectionResponse {
	return ConnectionResponse{
		Connection:  connection,
		Proof:       commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:   commitmenttypes.NewMerklePath(strings.Split(host.ConnectionPath(connectionID), "/")),
		ProofHeight: uint64(height),
	}
}

// QueryAllConnectionsParams defines the parameters necessary for querying for all
// connections.
type QueryAllConnectionsParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

// NewQueryAllConnectionsParams creates a new QueryAllConnectionsParams instance.
func NewQueryAllConnectionsParams(page, limit int) QueryAllConnectionsParams {
	return QueryAllConnectionsParams{
		Page:  page,
		Limit: limit,
	}
}

// ClientConnectionsResponse defines the client query response for a client
// connection paths which also includes a proof and the height from which the
// proof was retrieved.
type ClientConnectionsResponse struct {
	ConnectionPaths []string                    `json:"connection_paths" yaml:"connection_paths"`
	Proof           commitmenttypes.MerkleProof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath       commitmenttypes.MerklePath  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight     uint64                      `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewClientConnectionsResponse creates a new ConnectionPaths instance
func NewClientConnectionsResponse(
	clientID string, connectionPaths []string, proof *merkle.Proof, height int64,
) ClientConnectionsResponse {
	return ClientConnectionsResponse{
		ConnectionPaths: connectionPaths,
		Proof:           commitmenttypes.MerkleProof{Proof: proof},
		ProofPath:       commitmenttypes.NewMerklePath(strings.Split(host.ClientConnectionsPath(clientID), "/")),
		ProofHeight:     uint64(height),
	}
}

// QueryClientConnectionsParams defines the params for the following queries:
// - 'custom/ibc/client/<clientID>/connections'
type QueryClientConnectionsParams struct {
	ClientID string
}

// NewQueryClientConnectionsParams creates a new QueryClientConnectionsParams instance
func NewQueryClientConnectionsParams(clientID string) QueryClientConnectionsParams {
	return QueryClientConnectionsParams{
		ClientID: clientID,
	}
}
