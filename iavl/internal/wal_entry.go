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
}
