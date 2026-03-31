package blockstm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// BlockExecutionDebug holds all debugging/tracing data for a single block execution.
// It is safe for concurrent use by multiple executor goroutines.
type BlockExecutionDebug struct {
	mu   sync.Mutex
	txns []*TxnDebugData
}

// TxnDebugData captures all execution, validation, and suspension events for a single transaction.
type TxnDebugData struct {
	Executions  []ExecutionRecord  `json:"executions"`
	Validations []ValidationRecord `json:"validations"`
	Suspensions []SuspensionRecord `json:"suspensions"`
}

// ExecutionRecord captures a single execution attempt for a transaction.
type ExecutionRecord struct {
	Incarnation Incarnation          `json:"incarnation"`
	Start       time.Time            `json:"start"`
	End         time.Time            `json:"end"`
	ReadSets    map[int][]DebugRead  `json:"read_sets,omitempty"`
	WriteSets   map[int][]DebugWrite `json:"write_sets,omitempty"`
}

// DebugRead records a single key read during execution.
type DebugRead struct {
	Key          string `json:"key"`                    // hex-encoded key
	FromTxn      int    `json:"from_txn"`               // txn index the value was read from (-1 = storage)
	FromInc      uint   `json:"from_incarnation"`       // incarnation of the txn the value was read from
}

// DebugWrite records a single key written during execution.
type DebugWrite struct {
	Key      string `json:"key"`       // hex-encoded key
	ValueLen int    `json:"value_len"` // length of the value written (0 for deletes)
	IsDelete bool   `json:"is_delete"`
}

// ValidationRecord captures a single validation attempt for a transaction.
type ValidationRecord struct {
	Incarnation Incarnation `json:"incarnation"`
	Timestamp   time.Time   `json:"timestamp"`
	Valid       bool        `json:"valid"`
	Aborted     bool        `json:"aborted"`
}

// SuspensionRecord captures a single suspend/resume cycle for a transaction.
type SuspensionRecord struct {
	Suspend  time.Time `json:"suspend"`
	Resume   time.Time `json:"resume,omitempty"`
	BlockedBy TxnIndex `json:"blocked_by"`
}

func NewBlockExecutionDebug(blockSize int) *BlockExecutionDebug {
	txns := make([]*TxnDebugData, blockSize)
	for i := range txns {
		txns[i] = &TxnDebugData{}
	}
	return &BlockExecutionDebug{txns: txns}
}

// RecordExecution records an execution run with its read/write sets for the given transaction.
func (d *BlockExecutionDebug) RecordExecution(
	txn TxnIndex, incarnation Incarnation, start, end time.Time,
	reads map[int]*ReadSet, writes map[int][]WriteDescriptor,
) {
	d.mu.Lock()
	defer d.mu.Unlock()

	rec := ExecutionRecord{
		Incarnation: incarnation,
		Start:       start,
		End:         end,
	}

	if len(reads) > 0 {
		rec.ReadSets = make(map[int][]DebugRead, len(reads))
		for store, rs := range reads {
			var debugReads []DebugRead
			for _, rd := range rs.Reads {
				debugReads = append(debugReads, DebugRead{
					Key:     hex.EncodeToString(rd.Key),
					FromTxn: int(rd.Version.Index),
					FromInc: uint(rd.Version.Incarnation),
				})
			}
			for _, iter := range rs.Iterators {
				for _, rd := range iter.Reads {
					debugReads = append(debugReads, DebugRead{
						Key:     hex.EncodeToString(rd.Key),
						FromTxn: int(rd.Version.Index),
						FromInc: uint(rd.Version.Incarnation),
					})
				}
			}
			rec.ReadSets[store] = debugReads
		}
	}

	if len(writes) > 0 {
		rec.WriteSets = make(map[int][]DebugWrite, len(writes))
		for store, wds := range writes {
			debugWrites := make([]DebugWrite, len(wds))
			for i, wd := range wds {
				debugWrites[i] = DebugWrite{
					Key:      hex.EncodeToString(wd.Key),
					ValueLen: wd.ValueLen,
					IsDelete: wd.IsDelete,
				}
			}
			rec.WriteSets[store] = debugWrites
		}
	}

	d.txns[txn].Executions = append(d.txns[txn].Executions, rec)
}

// RecordValidation records a validation run for the given transaction.
func (d *BlockExecutionDebug) RecordValidation(txn TxnIndex, incarnation Incarnation, timestamp time.Time, valid, aborted bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
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
	d.mu.Lock()
	defer d.mu.Unlock()
	idx := len(d.txns[txn].Suspensions)
	d.txns[txn].Suspensions = append(d.txns[txn].Suspensions, SuspensionRecord{
		Suspend:   time.Now(),
		BlockedBy: blockedBy,
	})
	return idx
}

// RecordResume fills in the resume timestamp for a previously recorded suspension.
func (d *BlockExecutionDebug) RecordResume(txn TxnIndex, suspensionIdx int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if suspensionIdx < len(d.txns[txn].Suspensions) {
		d.txns[txn].Suspensions[suspensionIdx].Resume = time.Now()
	}
}

// Snapshot is the serializable form of all debug data for a block execution.
type Snapshot struct {
	BlockSize    int             `json:"block_size"`
	Transactions []*TxnDebugData `json:"transactions"`
}

// ToSnapshot returns a serializable snapshot of the debug data.
func (d *BlockExecutionDebug) ToSnapshot() *Snapshot {
	d.mu.Lock()
	defer d.mu.Unlock()
	// Deep copy the transaction data so the snapshot is independent.
	txns := make([]*TxnDebugData, len(d.txns))
	for i, t := range d.txns {
		cp := *t
		cp.Executions = make([]ExecutionRecord, len(t.Executions))
		copy(cp.Executions, t.Executions)
		cp.Validations = append([]ValidationRecord(nil), t.Validations...)
		cp.Suspensions = append([]SuspensionRecord(nil), t.Suspensions...)
		txns[i] = &cp
	}
	return &Snapshot{
		BlockSize:    len(d.txns),
		Transactions: txns,
	}
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
	d.mu.Lock()
	defer d.mu.Unlock()
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
