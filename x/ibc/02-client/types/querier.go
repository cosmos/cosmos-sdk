package types

// query routes supported by the IBC client Querier
const (
	QueryClientState    = "client_state"
	QueryConsensusState = "consensus_state"
	QueryVerifiedRoot   = "roots"
)

// QueryClientStateParams defines the params for the following queries:
// - 'custom/ibc/clients/<clientID>/client_state'
// - 'custom/ibc/clients/<clientID>/consensus_state'
type QueryClientStateParams struct {
	ClientID string
}

// NewQueryClientStateParams creates a new QueryClientStateParams instance
func NewQueryClientStateParams(id string) QueryClientStateParams {
	return QueryClientStateParams{
		ClientID: id,
	}
}

// QueryCommitmentRootParams defines the params for the following queries:
// - 'custom/ibc/clients/<clientID>/roots/<height>'
type QueryCommitmentRootParams struct {
	ClientID string
	Height   uint64
}

// NewQueryCommitmentRootParams creates a new QueryCommitmentRootParams instance
func NewQueryCommitmentRootParams(id string, height uint64) QueryCommitmentRootParams {
	return QueryCommitmentRootParams{
		ClientID: id,
		Height:   height,
	}
}
