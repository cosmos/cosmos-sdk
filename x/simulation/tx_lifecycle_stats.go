package simulation

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type TxLifecyclePhase string

const (
	TxPhaseCheckTx  TxLifecyclePhase = "checktx"
	TxPhasePrepare  TxLifecyclePhase = "prepare"
	TxPhaseProcess  TxLifecyclePhase = "process"
	TxPhaseFinalize TxLifecyclePhase = "finalize"
)

type txLifecycleStats struct {
	mu          sync.Mutex
	reasons     map[TxLifecyclePhase]map[string]int
	byMsgCounts map[TxLifecyclePhase]map[string]int
}

func newTxLifecycleStats() *txLifecycleStats {
	return &txLifecycleStats{
		reasons: map[TxLifecyclePhase]map[string]int{
			TxPhaseCheckTx:  {},
			TxPhasePrepare:  {},
			TxPhaseProcess:  {},
			TxPhaseFinalize: {},
		},
		byMsgCounts: map[TxLifecyclePhase]map[string]int{
			TxPhaseCheckTx:  {},
			TxPhasePrepare:  {},
			TxPhaseProcess:  {},
			TxPhaseFinalize: {},
		},
	}
}

func (s *txLifecycleStats) reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for phase := range s.reasons {
		s.reasons[phase] = map[string]int{}
		s.byMsgCounts[phase] = map[string]int{}
	}
}

func (s *txLifecycleStats) record(phase TxLifecyclePhase, reason string) {
	s.recordForMsg(phase, "", reason)
}

func (s *txLifecycleStats) recordForMsg(phase TxLifecyclePhase, msgType, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if strings.TrimSpace(msgType) != "" {
		s.byMsgCounts[phase][msgType]++
	}

	cleaned := strings.TrimSpace(reason)
	if cleaned == "" {
		cleaned = "unspecified"
	}
	if len(cleaned) > 200 {
		cleaned = cleaned[:200] + "..."
	}
	s.reasons[phase][cleaned]++
}

func (s *txLifecycleStats) summaryString() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	topN := SummaryTopNFromEnv()

	var b strings.Builder
	b.WriteString("Tx lifecycle failures:\n")
	for _, phase := range []TxLifecyclePhase{TxPhaseCheckTx, TxPhasePrepare, TxPhaseProcess, TxPhaseFinalize} {
		reasons := s.reasons[phase]
		msgCounts := s.byMsgCounts[phase]

		total := 0
		for _, count := range msgCounts {
			total += count
		}
		// Backward compatibility in case callers do not pass msgType.
		if total == 0 {
			for _, count := range reasons {
				total += count
			}
		}

		_, _ = fmt.Fprintf(&b, "- %s: %d\n", phase, total)

		if len(msgCounts) > 0 {
			msgEntries := ToSortedEntries(msgCounts)
			if topN > 0 && len(msgEntries) > topN {
				other := 0
				for _, e := range msgEntries[topN:] {
					other += e.Count
				}
				msgEntries = append(msgEntries[:topN], SummaryEntry{Key: "other", Count: other})
			}
			for _, e := range msgEntries {
				_, _ = fmt.Fprintf(&b, "  - %d\t%s\n", e.Count, e.Key)
			}
		}

		if len(reasons) == 0 {
			continue
		}

		reasonTotal := 0
		for _, count := range reasons {
			reasonTotal += count
		}
		if reasonTotal == 0 {
			continue
		}

		b.WriteString("    reasons:\n")
		reasonEntries := ToSortedEntries(reasons)
		if topN > 0 && len(reasonEntries) > topN {
			other := 0
			for _, e := range reasonEntries[topN:] {
				other += e.Count
			}
			reasonEntries = append(reasonEntries[:topN], SummaryEntry{Key: "other", Count: other})
		}
		for _, e := range reasonEntries {
			_, _ = fmt.Fprintf(&b, "      - %d\t%s\n", e.Count, e.Key)
		}
	}

	return b.String()
}

type txLifecycleStatsRegistry struct {
	mu    sync.Mutex
	stats map[uintptr]*txLifecycleStats
}

var simTxLifecycleStatsRegistry = txLifecycleStatsRegistry{
	stats: map[uintptr]*txLifecycleStats{},
}

func statsKeyForApp(app any) uintptr {
	if app == nil {
		return 0
	}

	v := reflect.ValueOf(app)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0
	}
	return v.Pointer()
}

func getTxLifecycleStatsForApp(app any) *txLifecycleStats {
	key := statsKeyForApp(app)

	simTxLifecycleStatsRegistry.mu.Lock()
	defer simTxLifecycleStatsRegistry.mu.Unlock()

	stats := simTxLifecycleStatsRegistry.stats[key]
	if stats == nil {
		stats = newTxLifecycleStats()
		simTxLifecycleStatsRegistry.stats[key] = stats
	}

	return stats
}

// ResetTxLifecycleFailuresForApp resets lifecycle stats for the given app.
func ResetTxLifecycleFailuresForApp(app any) {
	getTxLifecycleStatsForApp(app).reset()
}

// TxLifecycleFailuresSummary returns a per-phase failure summary for the
// simulation tx lifecycle.
func TxLifecycleFailuresSummary() string {
	return TxLifecycleFailuresSummaryForApp(nil)
}

// TxLifecycleFailuresSummaryForApp returns a per-phase failure summary for the
// simulation tx lifecycle scoped to a single app instance.
func TxLifecycleFailuresSummaryForApp(app any) string {
	return getTxLifecycleStatsForApp(app).summaryString()
}

// RecordTxLifecycleFailure records a lifecycle failure reason by phase.
func RecordTxLifecycleFailure(phase TxLifecyclePhase, reason string) {
	RecordTxLifecycleFailureForApp(nil, phase, reason)
}

// RecordTxLifecycleFailureForApp records a lifecycle failure reason by phase
// scoped to a single app instance.
func RecordTxLifecycleFailureForApp(app any, phase TxLifecyclePhase, reason string) {
	getTxLifecycleStatsForApp(app).record(phase, reason)
}

// RecordTxLifecycleFailureForMsg records a lifecycle failure reason by phase
// and message type.
func RecordTxLifecycleFailureForMsg(phase TxLifecyclePhase, msgType, reason string) {
	RecordTxLifecycleFailureForMsgForApp(nil, phase, msgType, reason)
}

// RecordTxLifecycleFailureForMsgForApp records a lifecycle failure reason by
// phase and message type scoped to a single app instance.
func RecordTxLifecycleFailureForMsgForApp(app any, phase TxLifecyclePhase, msgType, reason string) {
	getTxLifecycleStatsForApp(app).recordForMsg(phase, msgType, reason)
}

type TxLifecyclePhaseSummary struct {
	Total    int            `json:"total"`
	ByMsg    map[string]int `json:"by_msg"`
	ByReason map[string]int `json:"by_reason"`
}

type TxLifecycleSummarySnapshot struct {
	TotalRejected int                                          `json:"total_rejected"`
	Phases        map[TxLifecyclePhase]TxLifecyclePhaseSummary `json:"phases"`
}

func TxLifecycleFailuresSnapshot() TxLifecycleSummarySnapshot {
	return TxLifecycleFailuresSnapshotForApp(nil)
}

func TxLifecycleFailuresSnapshotForApp(app any) TxLifecycleSummarySnapshot {
	stats := getTxLifecycleStatsForApp(app)
	stats.mu.Lock()
	defer stats.mu.Unlock()

	phases := make(map[TxLifecyclePhase]TxLifecyclePhaseSummary, 4)
	totalRejected := 0

	for _, phase := range []TxLifecyclePhase{TxPhaseCheckTx, TxPhasePrepare, TxPhaseProcess, TxPhaseFinalize} {
		msgCounts := mapsCloneInt(stats.byMsgCounts[phase])
		reasons := mapsCloneInt(stats.reasons[phase])
		total := 0
		for _, c := range msgCounts {
			total += c
		}
		if total == 0 {
			for _, c := range reasons {
				total += c
			}
		}
		totalRejected += total
		phases[phase] = TxLifecyclePhaseSummary{
			Total:    total,
			ByMsg:    msgCounts,
			ByReason: reasons,
		}
	}

	return TxLifecycleSummarySnapshot{
		TotalRejected: totalRejected,
		Phases:        phases,
	}
}

func mapsCloneInt(src map[string]int) map[string]int {
	out := make(map[string]int, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
