package internal

type WALOpType int

const (
	WALOpSet WALOpType = iota
	WALOpDelete
	WALOpCommit
)

type WALEntry struct {
	Op                     WALOpType
	Version                uint64
	Key, Value             UnsafeBytes
	KeyOffset, ValueOffset int
	Checkpoint             bool
	// Offset is the offset of the start of this entry in the WAL file, which can be used for rollbacks
	Offset int
}
