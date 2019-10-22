package types

import (
	"strings"

	commitment "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
	"github.com/tendermint/tendermint/crypto/merkle"
)

// query routes supported by the IBC connection Querier
const (
	QueryConnection        = "connection"
	QueryClientConnections = "client_connections"
)

// ConnectionResponse defines the client query response for a connection which
// also includes a proof and the height from which the proof was retrieved.
type ConnectionResponse struct {
	Connection  ConnectionEnd    `json:"connection" yaml:"connection"`
	Proof       commitment.Proof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath   commitment.Path  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight uint64           `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewConnectionResponse creates a new ConnectionResponse instance
func NewConnectionResponse(
	connectionID string, connection ConnectionEnd, proof *merkle.Proof, height int64,
) ConnectionResponse {
	return ConnectionResponse{
		Connection:  connection,
		Proof:       commitment.Proof{Proof: proof},
		ProofPath:   commitment.NewPath(strings.Split(ConnectionPath(connectionID), "/")),
		ProofHeight: uint64(height),
	}
}

// QueryConnectionParams defines the params for the following queries:
// - 'custom/ibc/connections/<connectionID>'
type QueryConnectionParams struct {
	ConnectionID string
}

// NewQueryConnectionParams creates a new QueryConnectionParams instance
func NewQueryConnectionParams(clientID string) QueryConnectionParams {
	return QueryConnectionParams{
		ConnectionID: clientID,
	}
}

// ClientConnectionsResponse defines the client query response for a client
// connection paths which also includes a proof and the height from which the
// proof was retrieved.
type ClientConnectionsResponse struct {
	ConnectionPaths []string         `json:"connection_paths" yaml:"connection_paths"`
	Proof           commitment.Proof `json:"proof,omitempty" yaml:"proof,omitempty"`
	ProofPath       commitment.Path  `json:"proof_path,omitempty" yaml:"proof_path,omitempty"`
	ProofHeight     uint64           `json:"proof_height,omitempty" yaml:"proof_height,omitempty"`
}

// NewClientConnectionsResponse creates a new ConnectionPaths instance
func NewClientConnectionsResponse(
	clientID string, connectionPaths []string, proof *merkle.Proof, height int64,
) ClientConnectionsResponse {
	return ClientConnectionsResponse{
		ConnectionPaths: connectionPaths,
		Proof:           commitment.Proof{Proof: proof},
		ProofPath:       commitment.NewPath(strings.Split(ClientConnectionsPath(clientID), "/")),
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
