package simulation

import (
	"fmt"
	"os"
	"sort"
	"strconv"
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

type txLifecyclePhase = TxLifecyclePhase

const (
	txPhaseCheckTx  = TxPhaseCheckTx
	txPhasePrepare  = TxPhasePrepare
	txPhaseProcess  = TxPhaseProcess
	txPhaseFinalize = TxPhaseFinalize
)

type txLifecycleStats struct {
	mu          sync.Mutex
	reasons     map[TxLifecyclePhase]map[string]int
	byMsgCounts map[TxLifecyclePhase]map[string]int
}

func newTxLifecycleStats() *txLifecycleStats {
	return &txLifecycleStats{
		reasons: map[txLifecyclePhase]map[string]int{
			txPhaseCheckTx:  {},
			txPhasePrepare:  {},
			txPhaseProcess:  {},
			txPhaseFinalize: {},
		},
		byMsgCounts: map[txLifecyclePhase]map[string]int{
			txPhaseCheckTx:  {},
			txPhasePrepare:  {},
			txPhaseProcess:  {},
			txPhaseFinalize: {},
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
	topN := summaryTopNFromEnv()

	var b strings.Builder
	b.WriteString("Tx lifecycle failures:\n")
	for _, phase := range []TxLifecyclePhase{txPhaseCheckTx, txPhasePrepare, txPhaseProcess, txPhaseFinalize} {
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
			msgEntries := toSortedEntries(msgCounts)
			if topN > 0 && len(msgEntries) > topN {
				other := 0
				for _, e := range msgEntries[topN:] {
					other += e.Count
				}
				msgEntries = append(msgEntries[:topN], summaryEntry{Key: "other", Count: other})
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
		reasonEntries := toSortedEntries(reasons)
		if topN > 0 && len(reasonEntries) > topN {
			other := 0
			for _, e := range reasonEntries[topN:] {
				other += e.Count
			}
			reasonEntries = append(reasonEntries[:topN], summaryEntry{Key: "other", Count: other})
		}
		for _, e := range reasonEntries {
			_, _ = fmt.Fprintf(&b, "      - %d\t%s\n", e.Count, e.Key)
		}
	}

	return b.String()
}

var simTxLifecycleStats = newTxLifecycleStats()

// TxLifecycleFailuresSummary returns a per-phase failure summary for the
// simulation tx lifecycle.
func TxLifecycleFailuresSummary() string {
	return simTxLifecycleStats.summaryString()
}

// RecordTxLifecycleFailure records a lifecycle failure reason by phase.
func RecordTxLifecycleFailure(phase TxLifecyclePhase, reason string) {
	simTxLifecycleStats.record(phase, reason)
}

// RecordTxLifecycleFailureForMsg records a lifecycle failure reason by phase
// and message type.
func RecordTxLifecycleFailureForMsg(phase TxLifecyclePhase, msgType, reason string) {
	simTxLifecycleStats.recordForMsg(phase, msgType, reason)
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
	simTxLifecycleStats.mu.Lock()
	defer simTxLifecycleStats.mu.Unlock()

	phases := make(map[TxLifecyclePhase]TxLifecyclePhaseSummary, 4)
	totalRejected := 0

	for _, phase := range []TxLifecyclePhase{txPhaseCheckTx, txPhasePrepare, txPhaseProcess, txPhaseFinalize} {
		msgCounts := mapsCloneInt(simTxLifecycleStats.byMsgCounts[phase])
		reasons := mapsCloneInt(simTxLifecycleStats.reasons[phase])
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

func mapsCloneInt(src map[string]int) map[string]int {
	out := make(map[string]int, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}
