package blockstm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// BlockExecutionDebug holds all debugging/tracing data for a single block execution.
//
// Thread safety: the BlockSTM scheduler guarantees that at most one goroutine
// operates on a given transaction index at any time, so per-txn data requires no
// locking. The exported serialization methods (ToSnapshot, SaveToFile, etc.) must
// only be called after all executors have finished.
type BlockExecutionDebug struct {
	txns []*TxnDebugData
}

// TxnDebugData captures all execution, validation, and suspension events for a single transaction.
// Accessed by a single goroutine at a time (scheduler invariant).
type TxnDebugData struct {
	Executions  []ExecutionRecord
	Validations []ValidationRecord
	Suspensions []SuspensionRecord
}

// ExecutionRecord captures a single execution attempt for a transaction.
// Read/write sets are stored as raw references to avoid allocation and encoding
// on the hot execution path. Conversion to a serializable form happens only in
// ToSnapshot (cold path).
type ExecutionRecord struct {
	Incarnation Incarnation
	Start       time.Time
	End         time.Time
	Reads       map[int]*ReadSet
	Writes      map[int][]WriteDescriptor
}

// ValidationRecord captures a single validation attempt for a transaction.
type ValidationRecord struct {
	Incarnation Incarnation
	Timestamp   time.Time
	Valid       bool
	Aborted     bool
}

// SuspensionRecord captures a single suspend/resume cycle for a transaction.
type SuspensionRecord struct {
	Suspend   time.Time
	Resume    time.Time
	BlockedBy TxnIndex
}

func NewBlockExecutionDebug(blockSize int) *BlockExecutionDebug {
	txns := make([]*TxnDebugData, blockSize)
	for i := range txns {
		txns[i] = &TxnDebugData{}
	}
	return &BlockExecutionDebug{txns: txns}
}

// RecordExecution records an execution run for the given transaction.
// The reads and writes are stored by reference — no copying or encoding is done.
func (d *BlockExecutionDebug) RecordExecution(
	txn TxnIndex, incarnation Incarnation, start, end time.Time,
	reads map[int]*ReadSet, writes map[int][]WriteDescriptor,
) {
	d.txns[txn].Executions = append(d.txns[txn].Executions, ExecutionRecord{
		Incarnation: incarnation,
		Start:       start,
		End:         end,
		Reads:       reads,
		Writes:      writes,
	})
}

// RecordValidation records a validation run for the given transaction.
func (d *BlockExecutionDebug) RecordValidation(txn TxnIndex, incarnation Incarnation, timestamp time.Time, valid, aborted bool) {
	d.txns[txn].Validations = append(d.txns[txn].Validations, ValidationRecord{
		Incarnation: incarnation,
		Timestamp:   timestamp,
		Valid:       valid,
		Aborted:     aborted,
	})
}

// RecordSuspend records a suspension event and returns the index of the suspension record
// so the resume time can be filled in later.
func (d *BlockExecutionDebug) RecordSuspend(txn, blockedBy TxnIndex) int {
	idx := len(d.txns[txn].Suspensions)
	d.txns[txn].Suspensions = append(d.txns[txn].Suspensions, SuspensionRecord{
		Suspend:   time.Now(),
		BlockedBy: blockedBy,
	})
	return idx
}

// RecordResume fills in the resume timestamp for a previously recorded suspension.
func (d *BlockExecutionDebug) RecordResume(txn TxnIndex, suspensionIdx int) {
	if suspensionIdx < len(d.txns[txn].Suspensions) {
		d.txns[txn].Suspensions[suspensionIdx].Resume = time.Now()
	}
}

// --- Serializable types (JSON output, cold path only) ---

// Snapshot is the serializable form of all debug data for a block execution.
type Snapshot struct {
	BlockSize    int                `json:"block_size"`
	Transactions []*snapshotTxnData `json:"transactions"`
}

type snapshotTxnData struct {
	Executions  []snapshotExecution  `json:"executions"`
	Validations []snapshotValidation `json:"validations"`
	Suspensions []snapshotSuspension `json:"suspensions"`
}

type snapshotExecution struct {
	Incarnation Incarnation             `json:"incarnation"`
	Start       time.Time               `json:"start"`
	End         time.Time               `json:"end"`
	ReadSets    map[int][]snapshotRead  `json:"read_sets,omitempty"`
	WriteSets   map[int][]snapshotWrite `json:"write_sets,omitempty"`
}

type snapshotRead struct {
	Key     string `json:"key"`
	FromTxn int    `json:"from_txn"`
	FromInc uint   `json:"from_incarnation"`
}

type snapshotWrite struct {
	Key      string `json:"key"`
	ValueLen int    `json:"value_len"`
	IsDelete bool   `json:"is_delete"`
}

type snapshotValidation struct {
	Incarnation Incarnation `json:"incarnation"`
	Timestamp   time.Time   `json:"timestamp"`
	Valid       bool        `json:"valid"`
	Aborted     bool        `json:"aborted"`
}

type snapshotSuspension struct {
	Suspend   time.Time `json:"suspend"`
	Resume    time.Time `json:"resume,omitempty"`
	BlockedBy TxnIndex  `json:"blocked_by"`
}

// ToSnapshot converts the raw debug data into a serializable snapshot.
// All hex encoding and deep copying happens here, off the hot execution path.
// Must be called after all executors have finished.
func (d *BlockExecutionDebug) ToSnapshot() *Snapshot {
	txns := make([]*snapshotTxnData, len(d.txns))
	for i, t := range d.txns {
		st := &snapshotTxnData{
			Executions:  make([]snapshotExecution, len(t.Executions)),
			Validations: make([]snapshotValidation, len(t.Validations)),
			Suspensions: make([]snapshotSuspension, len(t.Suspensions)),
		}

		for j, exec := range t.Executions {
			se := snapshotExecution{
				Incarnation: exec.Incarnation,
				Start:       exec.Start,
				End:         exec.End,
			}
			if len(exec.Reads) > 0 {
				se.ReadSets = make(map[int][]snapshotRead, len(exec.Reads))
				for store, rs := range exec.Reads {
					reads := convertReads(rs)
					se.ReadSets[store] = reads
				}
			}
			if len(exec.Writes) > 0 {
				se.WriteSets = make(map[int][]snapshotWrite, len(exec.Writes))
				for store, wds := range exec.Writes {
					writes := make([]snapshotWrite, len(wds))
					for k, wd := range wds {
						writes[k] = snapshotWrite{
							Key:      hex.EncodeToString(wd.Key),
							ValueLen: wd.ValueLen,
							IsDelete: wd.IsDelete,
						}
					}
					se.WriteSets[store] = writes
				}
			}
			st.Executions[j] = se
		}

		for j, val := range t.Validations {
			st.Validations[j] = snapshotValidation(val)
		}

		for j, sus := range t.Suspensions {
			st.Suspensions[j] = snapshotSuspension(sus)
		}

		txns[i] = st
	}
	return &Snapshot{
		BlockSize:    len(d.txns),
		Transactions: txns,
	}
}

func convertReads(rs *ReadSet) []snapshotRead {
	n := len(rs.Reads)
	for _, iter := range rs.Iterators {
		n += len(iter.Reads)
	}
	reads := make([]snapshotRead, 0, n)
	for _, rd := range rs.Reads {
		reads = append(reads, snapshotRead{
			Key:     hex.EncodeToString(rd.Key),
			FromTxn: int(rd.Version.Index),
			FromInc: uint(rd.Version.Incarnation),
		})
	}
	for _, iter := range rs.Iterators {
		for _, rd := range iter.Reads {
			reads = append(reads, snapshotRead{
				Key:     hex.EncodeToString(rd.Key),
				FromTxn: int(rd.Version.Index),
				FromInc: uint(rd.Version.Incarnation),
			})
		}
	}
	return reads
}

// MarshalJSON serializes the debug data to JSON.
func (d *BlockExecutionDebug) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.ToSnapshot())
}

// SaveToFile writes the debug data to a file in JSON format.
func (d *BlockExecutionDebug) SaveToFile(path string) error {
	data, err := json.MarshalIndent(d.ToSnapshot(), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal debug data: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// LoadSnapshot reads a Snapshot from a JSON file.
func LoadSnapshot(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read debug file: %w", err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return nil, fmt.Errorf("unmarshal debug data: %w", err)
	}
	return &snap, nil
}

// DumpSequencing prints a human-readable summary of the execution timeline.
func (d *BlockExecutionDebug) DumpSequencing() {
	for i, txn := range d.txns {
		if len(txn.Executions) == 0 && len(txn.Validations) == 0 {
			fmt.Printf("txn %d: no data\n", i)
			continue
		}
		fmt.Printf("txn %d:\n", i)
		for _, exec := range txn.Executions {
			fmt.Printf("  exec incarnation=%d start=%v end=%v duration=%v\n",
				exec.Incarnation, exec.Start, exec.End, exec.End.Sub(exec.Start))
		}
		for _, val := range txn.Validations {
			fmt.Printf("  validate incarnation=%d at=%v valid=%v aborted=%v\n",
				val.Incarnation, val.Timestamp, val.Valid, val.Aborted)
		}
		for _, sus := range txn.Suspensions {
			fmt.Printf("  suspended blocked_by=%d at=%v resumed=%v\n",
				sus.BlockedBy, sus.Suspend, sus.Resume)
		}
	}
}
