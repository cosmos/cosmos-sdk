package types

// Querier routes for the evidence module
const (
	QueryEvidence    = "evidence"
	QueryAllEvidence = "all_evidence"
)

// QueryEvidenceParams defines the parameters necessary for querying Evidence.
type QueryEvidenceParams struct {
	EvidenceHash string
}

func NewQueryEvidenceParams(hash string) QueryEvidenceParams {
	return QueryEvidenceParams{EvidenceHash: hash}
}

type QueryAllEvidenceParams struct {
	Page  int
	Limit int
}

func NewQueryAllEvidenceParams(page, limit int) QueryAllEvidenceParams {
	return QueryAllEvidenceParams{Page: page, Limit: limit}
}
