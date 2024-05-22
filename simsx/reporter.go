package simsx

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/gogoproto/proto"
)

type ReportResult struct {
	Error      error
	MsgProtoBz []byte
}

type SimulationReporter interface {
	WithScope(msg sdk.Msg) SimulationReporter
	WithT(t testing.TB) SimulationReporter
	Skip(comment string)
	Skipf(comment string, args ...any)
	IsSkipped() bool
	ToLegacyOperationMsg() simtypes.OperationMsg
	// complete with failure
	Fail(err error, comments ...string)
	// complete with success
	Success(msg sdk.Msg, comments ...string)
	// error captured on fail
	ExecutionResult() ReportResult
	Comment() string
}

var _ SimulationReporter = &BasicSimulationReporter{}

type ReporterStatus uint8

const (
	undefined ReporterStatus = iota
	skipped   ReporterStatus = iota
	completed ReporterStatus = iota
)

type BasicSimulationReporter struct {
	t          testing.TB
	module     string
	msgTypeURL string
	error      error
	comments   []string
	status     ReporterStatus
	msgProtoBz []byte
}

func NewBasicSimulationReporter(t testing.TB) *BasicSimulationReporter {
	return &BasicSimulationReporter{t: t}
}

func (x BasicSimulationReporter) WithScope(msg sdk.Msg) SimulationReporter {
	typeURL := sdk.MsgTypeURL(msg)
	return &BasicSimulationReporter{
		t:          x.t,
		error:      x.error,
		status:     x.status,
		msgProtoBz: x.msgProtoBz,
		msgTypeURL: typeURL,
		module:     sdk.GetModuleNameFromTypeURL(typeURL),
		comments:   slices.Clone(x.comments),
	}
}

func (x BasicSimulationReporter) WithT(t testing.TB) SimulationReporter {
	return &BasicSimulationReporter{
		t:          t,
		error:      x.error,
		status:     x.status,
		msgProtoBz: x.msgProtoBz,
		msgTypeURL: x.msgTypeURL,
		module:     x.module,
		comments:   slices.Clone(x.comments),
	}
}

func (x *BasicSimulationReporter) Skip(comment string) {
	x.toStatus(skipped, comment)
}

func (x *BasicSimulationReporter) Skipf(comment string, args ...any) {
	x.Skip(fmt.Sprintf(comment, args...))
}

func (x BasicSimulationReporter) IsSkipped() bool {
	return x.status > undefined
}

func (x *BasicSimulationReporter) ToLegacyOperationMsg() simtypes.OperationMsg {
	switch x.status {
	case skipped:
		return simtypes.NoOpMsg(x.module, x.msgTypeURL, x.Comment())
	case completed:
		if x.error == nil {
			return simtypes.NoOpMsg(x.module, x.msgTypeURL, x.Comment())
		} else {
			return simtypes.NewOperationMsgBasic(x.module, x.msgTypeURL, x.Comment(), true, x.msgProtoBz)
		}
	default:
		x.Fail(errors.New("operation aborted before msg was executed"))
		return x.ToLegacyOperationMsg()
	}
}

func (x *BasicSimulationReporter) Fail(err error, comments ...string) {
	x.toStatus(completed, comments...)
	x.error = err
}

func (x *BasicSimulationReporter) Success(msg sdk.Msg, comments ...string) {
	x.toStatus(completed, comments...)
	protoBz, err := proto.Marshal(msg) // todo: not great to capture the proto bytes here again but legacy test are using it.
	if err != nil {
		panic(err)
	}
	x.msgProtoBz = protoBz
}

func (x BasicSimulationReporter) ExecutionResult() ReportResult {
	return ReportResult{Error: x.error, MsgProtoBz: x.msgProtoBz}
}

func (x *BasicSimulationReporter) toStatus(next ReporterStatus, comments ...string) {
	if x.status > next {
		panic(fmt.Sprintf("can not switch from status %d to %d", x.status, next))
	}
	x.status = next
	x.comments = append(x.comments, comments...)
	if x.t != nil && x.status == skipped {
		x.t.Skip(x.Comment())
	}
}

func (x BasicSimulationReporter) Comment() string {
	return strings.Join(x.comments, ", ")
}
