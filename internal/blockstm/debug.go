package blockstm

import (
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
	Incarnation Incarnation `json:"incarnation"`
	Start       time.Time   `json:"start"`
	End         time.Time   `json:"end"`
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

// RecordExecution records an execution run for the given transaction.
func (d *BlockExecutionDebug) RecordExecution(txn TxnIndex, incarnation Incarnation, start, end time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.txns[txn].Executions = append(d.txns[txn].Executions, ExecutionRecord{
		Incarnation: incarnation,
		Start:       start,
		End:         end,
	})
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
		cp.Executions = append([]ExecutionRecord(nil), t.Executions...)
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
