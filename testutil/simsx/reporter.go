package simsx

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// SimulationReporter is an interface for reporting the result of a simulation run.
type SimulationReporter interface {
	WithScope(msg sdk.Msg, optionalSkipHook ...SkipHook) SimulationReporter
	Skip(comment string)
	Skipf(comment string, args ...any)
	// IsSkipped returns true when skipped or completed
	IsSkipped() bool
	ToLegacyOperationMsg() simtypes.OperationMsg
	// Fail complete with failure
	Fail(err error, comments ...string)
	// Success complete with success
	Success(msg sdk.Msg, comments ...string)
	// Close returns error captured on fail
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

func reporterStatusFrom(s uint32) ReporterStatus {
	switch s {
	case uint32(skipped):
		return skipped
	case uint32(completed):
		return completed
	default:
		return undefined
	}
}

// SkipHook is an interface that represents a callback hook used triggered on skip operations.
// It provides a single method `Skip` that accepts variadic arguments. This interface is implemented
// by Go stdlib testing.T and testing.B
type SkipHook interface {
	Skip(args ...any)
}

var _ SkipHook = SkipHookFn(nil)

type SkipHookFn func(args ...any)

func (s SkipHookFn) Skip(args ...any) {
	s(args...)
}

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
		r.summary.Add(child.module, child.msgTypeURL, reporterStatusFrom(child.status.Load()), child.Comment())
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
func (x *BasicSimulationReporter) WithScope(msg sdk.Msg, optionalSkipHook ...SkipHook) SimulationReporter {
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
	x.toStatus(skipped, comment)
}

func (x *BasicSimulationReporter) Skipf(comment string, args ...any) {
	x.Skip(fmt.Sprintf(comment, args...))
}

func (x *BasicSimulationReporter) IsSkipped() bool {
	return reporterStatusFrom(x.status.Load()) > undefined
}

func (x *BasicSimulationReporter) ToLegacyOperationMsg() simtypes.OperationMsg {
	switch reporterStatusFrom(x.status.Load()) {
	case skipped:
		return simtypes.NoOpMsg(x.module, x.msgTypeURL, x.Comment())
	case completed:
		x.cMX.RLock()
		err := x.error
		x.cMX.RUnlock()
		if err == nil {
			return simtypes.NewOperationMsgBasic(x.module, x.msgTypeURL, x.Comment(), true, []byte{})
		} else {
			return simtypes.NewOperationMsgBasic(x.module, x.msgTypeURL, x.Comment(), false, []byte{})
		}
	default:
		x.Fail(errors.New("operation aborted before msg was executed"))
		return x.ToLegacyOperationMsg()
	}
}

func (x *BasicSimulationReporter) Fail(err error, comments ...string) {
	if !x.toStatus(completed, comments...) {
		return
	}
	x.cMX.Lock()
	defer x.cMX.Unlock()
	x.error = err
}

func (x *BasicSimulationReporter) Success(msg sdk.Msg, comments ...string) {
	if !x.toStatus(completed, comments...) {
		return
	}
	if msg == nil {
		return
	}
}

func (x *BasicSimulationReporter) Close() error {
	x.completedCallback(x)
	x.cMX.RLock()
	defer x.cMX.RUnlock()
	return x.error
}

func (x *BasicSimulationReporter) toStatus(next ReporterStatus, comments ...string) bool {
	oldStatus := reporterStatusFrom(x.status.Load())
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

	if oldStatus != skipped && next == skipped {
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
	if status == completed {
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
	topN := summaryTopNFromEnv()
	keys := slices.Sorted(maps.Keys(s.counts))
	var sb strings.Builder
	for _, key := range keys {
		fmt.Fprintf(&sb, "%s: %d\n", key, s.counts[key])
	}
	if len(s.skipReasons) != 0 {
		sb.WriteString("\nSkip reasons:\n")
	}
	msgTypeTotals := make(map[string]int, len(s.skipReasons))
	for msgType, reasons := range s.skipReasons {
		msgTypeTotals[msgType] = sum(slices.Collect(maps.Values(reasons)))
	}
	msgTypeEntries := toSortedEntries(msgTypeTotals)
	if topN > 0 && len(msgTypeEntries) > topN {
		other := 0
		for _, e := range msgTypeEntries[topN:] {
			other += e.Count
		}
		msgTypeEntries = append(msgTypeEntries[:topN], summaryEntry{Key: "other", Count: other})
	}
	for _, msgTypeEntry := range msgTypeEntries {
		m := msgTypeEntry.Key
		if m == "other" {
			fmt.Fprintf(&sb, "%d\t%s\n", msgTypeEntry.Count, "other")
			continue
		}
		c := s.skipReasons[m]
		values := maps.Values(c)
		total := sum(slices.Collect(values))
		reasonEntries := toSortedEntries(c)
		if topN > 0 && len(reasonEntries) > topN {
			other := 0
			for _, e := range reasonEntries[topN:] {
				other += e.Count
			}
			reasonEntries = append(reasonEntries[:topN], summaryEntry{Key: "other", Count: other})
		}
		reasons := make([]string, 0, len(reasonEntries))
		for _, e := range reasonEntries {
			reasons = append(reasons, e.Key)
		}
		fmt.Fprintf(&sb, "%d\t%s: %q\n", total, m, reasons)
	}
	return sb.String()
}

func (s *ExecutionSummary) TotalSkipped() int {
	s.mx.RLock()
	defer s.mx.RUnlock()
	total := 0
	for key, count := range s.counts {
		if strings.HasSuffix(key, "_skipped") {
			total += count
		}
	}
	return total
}

func (s *ExecutionSummary) TotalCompleted() int {
	s.mx.RLock()
	defer s.mx.RUnlock()
	total := 0
	for key, count := range s.counts {
		if strings.HasSuffix(key, "_completed") {
			total += count
		}
	}
	return total
}

type SummarySnapshot struct {
	Counts         map[string]int            `json:"counts"`
	SkipReasons    map[string]map[string]int `json:"skip_reasons"`
	TotalSkipped   int                       `json:"total_skipped"`
	TotalCompleted int                       `json:"total_completed"`
}

func (s *ExecutionSummary) Snapshot() SummarySnapshot {
	s.mx.RLock()
	defer s.mx.RUnlock()

	counts := maps.Clone(s.counts)
	reasons := make(map[string]map[string]int, len(s.skipReasons))
	for msgType, byReason := range s.skipReasons {
		reasons[msgType] = maps.Clone(byReason)
	}

	skipped := 0
	completed := 0
	for key, count := range counts {
		switch {
		case strings.HasSuffix(key, "_skipped"):
			skipped += count
		case strings.HasSuffix(key, "_completed"):
			completed += count
		}
	}

	return SummarySnapshot{
		Counts:         counts,
		SkipReasons:    reasons,
		TotalSkipped:   skipped,
		TotalCompleted: completed,
	}
}

type summaryEntry struct {
	Key   string
	Count int
}

func toSortedEntries(items map[string]int) []summaryEntry {
	entries := make([]summaryEntry, 0, len(items))
	for k, c := range items {
		entries = append(entries, summaryEntry{Key: k, Count: c})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count == entries[j].Count {
			return entries[i].Key < entries[j].Key
		}
		return entries[i].Count > entries[j].Count
	})
	return entries
}

func summaryTopNFromEnv() int {
	v := strings.TrimSpace(os.Getenv("SIMAPP_SUMMARY_TOP_N"))
	if v == "" {
		return 0
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0
	}
	return n
}

func sum(values []int) int {
	var r int
	for _, v := range values {
		r += v
	}
	return r
}
