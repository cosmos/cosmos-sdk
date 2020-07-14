package types

// query endpoints supported by the upgrade Querier
const (
	QueryCurrent = "current"
	QueryApplied = "applied"
)

// NewQueryCurrentPlanRequest creates a new instance of QueryCurrentPlanRequest.
func NewQueryCurrentPlanRequest() *QueryCurrentPlanRequest {
	return &QueryCurrentPlanRequest{}
}

// NewQueryAppliedPlanRequest creates a new instance of QueryAppliedPlanRequest.
func NewQueryAppliedPlanRequest(name string) *QueryAppliedPlanRequest {
	return &QueryAppliedPlanRequest{Name: name}
}
