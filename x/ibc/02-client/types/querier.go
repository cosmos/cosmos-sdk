package types

// query routes supported by the IBC client Querier
const (
	QueryClientState    = "clientState"
	QueryConsensusState = "consensusState"
	QueryCommitmentPath = "commitmentPath"
	QueryCommitmentRoot = "commitmentRoot"
	QueryHeader         = "header"
)

// QueryClientStateParams defines the params for the following queries:
// - 'custom/ibc/client/clientState'
type QueryClientStateParams struct {
	ID string
}

// NewQueryClientStateParams creates a new QueryClientStateParams instance
func NewQueryClientStateParams(id string) QueryClientStateParams {
	return QueryClientStateParams{
		ID: id,
	}
}

// QueryCommitmentRootParams defines the params for the following queries:
// - 'custom/ibc/client/commitmentRoot'
type QueryCommitmentRootParams struct {
	ID     string
	Height uint64
}

// NewQueryCommitmentRootParams creates a new QueryCommitmentRootParams instance
func NewQueryCommitmentRootParams(id string, height uint64) QueryCommitmentRootParams {
	return QueryCommitmentRootParams{
		ID:     id,
		Height: height,
	}
}
