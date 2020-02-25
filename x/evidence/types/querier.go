package types

// Querier routes for the evidence module
const (
	QueryParameters  = "parameters"
	QueryEvidence    = "evidence"
	QueryAllEvidence = "all_evidence"
)

// QueryEvidenceParams defines the parameters necessary for querying Evidence.
type QueryEvidenceParams struct {
	EvidenceHash string `json:"evidence_hash" yaml:"evidence_hash"`
}

func NewQueryEvidenceParams(hash string) QueryEvidenceParams {
	return QueryEvidenceParams{EvidenceHash: hash}
}

// QueryAllEvidenceParams defines the parameters necessary for querying for all Evidence.
type QueryAllEvidenceParams struct {
	Page  int `json:"page" yaml:"page"`
	Limit int `json:"limit" yaml:"limit"`
}

func NewQueryAllEvidenceParams(page, limit int) QueryAllEvidenceParams {
	return QueryAllEvidenceParams{Page: page, Limit: limit}
}
