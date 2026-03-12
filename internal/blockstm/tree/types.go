package tree

import "bytes"

const (
	TelemetrySubsystem = "blockstm_btree"

	KeyGet          = "get"
	KeySet          = "set"
	KeyDelete       = "delete"
	KeyReverseSeek  = "reverse_seek"
	KeyScan         = "scan"
	KeyGetOrDefault = "get_or_default"
)

type KeyItem interface {
	GetKey() []byte
}

func KeyItemLess[T KeyItem](a, b T) bool {
	return bytes.Compare(a.GetKey(), b.GetKey()) < 0
}
