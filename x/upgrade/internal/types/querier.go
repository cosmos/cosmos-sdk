package types

// query endpoints supported by the upgrade Querier
const (
	QueryCurrent = "current"
	QueryApplied = "applied"
)

// QueryAppliedParams is passed as data with QueryApplied
type QueryAppliedParams struct {
	Name string
}

// NewQueryAppliedParams creates a new instance to query
// if a named plan was applied
func NewQueryAppliedParams(name string) QueryAppliedParams {
	return QueryAppliedParams{Name: name}
}
