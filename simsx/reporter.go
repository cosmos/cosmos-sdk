package simsx

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"golang.org/x/exp/maps"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/gogoproto/proto"
)

type ReportResult struct {
	Status     string
	Error      error
	MsgProtoBz []byte
}

func (r ReportResult) String() string {
	return fmt.Sprintf("error: %q, status: %q", r.Error, r.Status)
}

type SimulationReporter interface {
	WithScope(msg sdk.Msg) SimulationReporter
	Skip(comment string)
	Skipf(comment string, args ...any)
	IsSkipped() bool
	ToLegacyOperationMsg() simtypes.OperationMsg
	// complete with failure
	Fail(err error, comments ...string)
	// complete with success
	Success(msg sdk.Msg, comments ...string)
	// error captured on fail
	Close() error
	Comment() string
}

var _ SimulationReporter = &BasicSimulationReporter{}

type ReporterStatus uint8

const (
	undefined ReporterStatus = iota
	skipped   ReporterStatus = iota
	completed ReporterStatus = iota
)

func (s ReporterStatus) String() string {
	switch s {
	case skipped:
		return "skipped"
	case completed:
		return "completed"
	default:
		return "undefined"
	}
}

type SkipHook interface {
	Skip(args ...any)
}
type BasicSimulationReporter struct {
	skipCallbacks     []SkipHook
	completedCallback func(reporter BasicSimulationReporter)
	summary           *ExecutionSummary
	module            string
	msgTypeURL        string
	error             error
	comments          []string
	status            ReporterStatus
	msgProtoBz        []byte
}

// NewBasicSimulationReporter constructor that accepts an optional callback hook that is called on state transition to skipped status
// A typical implementation for this hook is testing.T
func NewBasicSimulationReporter(optionalSkipHook ...SkipHook) *BasicSimulationReporter {
	r := &BasicSimulationReporter{
		skipCallbacks: optionalSkipHook,
		summary:       NewExecutionSummary(),
	}
	r.completedCallback = func(child BasicSimulationReporter) {
		r.summary.Add(child.module, child.msgTypeURL, child.status, child.Comment())
	}
	return r
}

func (x *BasicSimulationReporter) WithScope(msg sdk.Msg) SimulationReporter {
	typeURL := sdk.MsgTypeURL(msg)
	return &BasicSimulationReporter{
		skipCallbacks:     x.skipCallbacks,
		completedCallback: x.completedCallback,
		error:             x.error,
		status:            x.status,
		msgProtoBz:        x.msgProtoBz,
		msgTypeURL:        typeURL,
		module:            sdk.GetModuleNameFromTypeURL(typeURL),
		comments:          slices.Clone(x.comments),
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

func (x BasicSimulationReporter) Close() error {
	x.completedCallback(x)
	return x.error
	// return ReportResult{Error: x.error, MsgProtoBz: x.msgProtoBz, Status: x.status.String()}
}

func (x *BasicSimulationReporter) toStatus(next ReporterStatus, comments ...string) {
	if x.status > next {
		panic(fmt.Sprintf("can not switch from status %d to %d", x.status, next))
	}
	x.comments = append(x.comments, comments...)
	if x.status != skipped && next == skipped {
		for _, hook := range x.skipCallbacks {
			hook.Skip(x.Comment())
		}
	}
	x.status = next
}

func (x BasicSimulationReporter) Comment() string {
	return strings.Join(x.comments, ", ")
}

func (x BasicSimulationReporter) Summary() ExecutionSummary {
	return *x.summary
}

type ExecutionSummary struct {
	counts  map[string]int
	reasons map[string]map[string]struct{}
}

func NewExecutionSummary() *ExecutionSummary {
	return &ExecutionSummary{counts: make(map[string]int), reasons: make(map[string]map[string]struct{})}
}

func (s *ExecutionSummary) Add(module string, url string, status ReporterStatus, comment string) {
	combinedKey := fmt.Sprintf("%s_%s", module, status.String())
	s.counts[combinedKey] += 1
	if status == completed {
		return
	}
	r, ok := s.reasons[url]
	if !ok {
		r = make(map[string]struct{})
		s.reasons[url] = r
	}
	r[comment] = struct{}{}
}

func (s ExecutionSummary) String() string {
	keys := maps.Keys(s.counts)
	sort.Strings(keys)
	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(fmt.Sprintf("%s: %d\n", key, s.counts[key]))
	}
	for m, c := range s.reasons {
		sb.WriteString(fmt.Sprintf("%s: %q\n", m, c))
	}
	return sb.String()
}
