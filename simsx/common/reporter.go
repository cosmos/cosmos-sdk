package common

import (
	"fmt"
	"maps"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ SimulationReporter = &BasicSimulationReporter{}

var _ SkipHook = SkipHookFn(nil)

type BasicSimulationReporter struct {
	skipCallbacks     []SkipHook
	completedCallback func(reporter *BasicSimulationReporter)
	module            string
	msgTypeURL        string

	status atomic.Uint32

	cMX      sync.RWMutex
	comments []string
	error    error

	summary *ExecutionSummary
}

// NewBasicSimulationReporter constructor that accepts an optional callback hook that is called on state transition to skipped status
// A typical implementation for this hook is testing.T or testing.B.
func NewBasicSimulationReporter(optionalSkipHook ...SkipHook) *BasicSimulationReporter {
	r := &BasicSimulationReporter{
		skipCallbacks: optionalSkipHook,
		summary:       NewExecutionSummary(),
	}
	r.completedCallback = func(child *BasicSimulationReporter) {
		r.summary.Add(child.module, child.msgTypeURL, ReporterStatus(child.status.Load()), child.Comment())
	}
	return r
}

// WithScope is a method of the BasicSimulationReporter type that creates a new instance of SimulationReporter
// with an additional scope specified by the input `msg`. The msg is used to set type, module and binary data as
// context for the legacy operation.
// The WithScope method acts as a constructor to initialize state and has to be called before using the instance
// in DeliverSimsMsg.
//
// The method accepts an optional `optionalSkipHook` parameter
// that can be used to add a callback hook that is triggered on skip operations additional to any parent skip hook.
// This method returns the newly created
// SimulationReporter instance.
func (x *BasicSimulationReporter) WithScope(msg sdk.Msg, optionalSkipHook ...SkipHook) SimulationReporterRuntime {
	typeURL := sdk.MsgTypeURL(msg)
	r := &BasicSimulationReporter{
		skipCallbacks:     append(x.skipCallbacks, optionalSkipHook...),
		completedCallback: x.completedCallback,
		error:             x.error,
		msgTypeURL:        typeURL,
		module:            sdk.GetModuleNameFromTypeURL(typeURL),
		comments:          slices.Clone(x.comments),
	}
	r.status.Store(x.status.Load())
	return r
}

func (x *BasicSimulationReporter) Skip(comment string) {
	x.toStatus(ReporterStatusSkipped, comment)
}

func (x *BasicSimulationReporter) Skipf(comment string, args ...any) {
	x.Skip(fmt.Sprintf(comment, args...))
}

func (x *BasicSimulationReporter) IsAborted() bool {
	return ReporterStatus(x.status.Load()) > ReporterStatusUndefined
}

func (x *BasicSimulationReporter) Scope() (string, string) {
	return x.module, x.msgTypeURL
}

func (x *BasicSimulationReporter) Fail(err error, comments ...string) {
	if !x.toStatus(ReporterStatusCompleted, comments...) {
		return
	}
	x.cMX.Lock()
	defer x.cMX.Unlock()
	x.error = err
}

func (x *BasicSimulationReporter) Success(msg sdk.Msg, comments ...string) {
	if !x.toStatus(ReporterStatusCompleted, comments...) {
		return
	}
	if msg == nil {
		return
	}
}

func (x *BasicSimulationReporter) Error() error {
	x.cMX.RLock()
	defer x.cMX.RUnlock()
	return x.error
}
func (x *BasicSimulationReporter) Close() error {
	x.completedCallback(x)
	return x.Error()
}

// IsSkipped
// Deprecated: use IsAborted instead
func (x *BasicSimulationReporter) IsSkipped() bool {
	return x.IsAborted()
}

func (x *BasicSimulationReporter) Status() ReporterStatus {
	return ReporterStatus(x.status.Load())
}

// transition to next status
func (x *BasicSimulationReporter) toStatus(next ReporterStatus, comments ...string) bool {
	oldStatus := ReporterStatus(x.status.Load())
	if oldStatus > next {
		panic(fmt.Sprintf("can not switch from status %s to %s", oldStatus, next))
	}
	if !x.status.CompareAndSwap(uint32(oldStatus), uint32(next)) {
		return false
	}
	x.cMX.Lock()
	newComments := append(x.comments, comments...)
	x.comments = newComments
	x.cMX.Unlock()

	if oldStatus != ReporterStatusSkipped && next == ReporterStatusSkipped {
		prettyComments := strings.Join(newComments, ", ")
		for _, hook := range x.skipCallbacks {
			hook.Skip(prettyComments)
		}
	}
	return true
}

func (x *BasicSimulationReporter) Comment() string {
	x.cMX.RLock()
	defer x.cMX.RUnlock()
	return strings.Join(x.comments, ", ")
}

func (x *BasicSimulationReporter) Summary() *ExecutionSummary {
	return x.summary
}

type ExecutionSummary struct {
	mx          sync.RWMutex
	counts      map[string]int            // module to count
	skipReasons map[string]map[string]int // msg type to reason->count
}

func NewExecutionSummary() *ExecutionSummary {
	return &ExecutionSummary{counts: make(map[string]int), skipReasons: make(map[string]map[string]int)}
}

func (s *ExecutionSummary) Add(module, url string, status ReporterStatus, comment string) {
	s.mx.Lock()
	defer s.mx.Unlock()
	combinedKey := fmt.Sprintf("%s_%s", module, status.String())
	s.counts[combinedKey] += 1
	if status == ReporterStatusCompleted {
		return
	}
	r, ok := s.skipReasons[url]
	if !ok {
		r = make(map[string]int)
		s.skipReasons[url] = r
	}
	r[comment] += 1
}

func (s *ExecutionSummary) String() string {
	s.mx.RLock()
	defer s.mx.RUnlock()
	keys := slices.Sorted(maps.Keys(s.counts))
	var sb strings.Builder
	for _, key := range keys {
		sb.WriteString(fmt.Sprintf("%s: %d\n", key, s.counts[key]))
	}
	if len(s.skipReasons) != 0 {
		sb.WriteString("\nSkip reasons:\n")
	}
	for m, c := range s.skipReasons {
		values := maps.Values(c)
		keys := maps.Keys(c)
		sb.WriteString(fmt.Sprintf("%d\t%s: %q\n", sum(slices.Collect(values)), m, slices.Collect(keys)))
	}
	return sb.String()
}

func sum(values []int) int {
	var r int
	for _, v := range values {
		r += v
	}
	return r
}
