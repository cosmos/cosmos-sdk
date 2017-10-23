package sdk

import (
	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/go-wire/data"
)

//---------- results and some wrappers --------

// Result is a common interface of CheckResult and GetResult
type Result interface {
	GetData() data.Bytes
	GetLog() string
}

func ToABCI(r Result) abci.Result {
	return abci.Result{
		Data: r.GetData(),
		Log:  r.GetLog(),
	}
}

// CheckResult captures any non-error abci result
// to make sure people use error for error cases
type CheckResult struct {
	Data data.Bytes
	Log  string
	// GasAllocated is the maximum units of work we allow this tx to perform
	GasAllocated uint64
	// GasPayment is the total fees for this tx (or other source of payment)
	GasPayment uint64
}

// NewCheck sets the gas used and the response data but no more info
// these are the most common info needed to be set by the Handler
func NewCheck(gasAllocated uint64, log string) CheckResult {
	return CheckResult{
		GasAllocated: gasAllocated,
		Log:          log,
	}
}

var _ Result = CheckResult{}

func (r CheckResult) GetData() data.Bytes {
	return r.Data
}

func (r CheckResult) GetLog() string {
	return r.Log
}

// DeliverResult captures any non-error abci result
// to make sure people use error for error cases
type DeliverResult struct {
	Data    data.Bytes
	Log     string
	Diff    []*abci.Validator
	GasUsed uint64
}

var _ Result = DeliverResult{}

func (r DeliverResult) GetData() data.Bytes {
	return r.Data
}

func (r DeliverResult) GetLog() string {
	return r.Log
}
