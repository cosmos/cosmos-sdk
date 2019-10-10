package types

// query routes supported by the IBC connection Querier
const (
	QueryClientState    = "clientState"
	QueryConsensusState = "consensusState"
	QueryCommitmentPath = "commitmentPath"
	QueryCommitmentRoot = "roots"
)

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
