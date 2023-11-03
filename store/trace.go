package store

import "golang.org/x/exp/maps"

// TraceContext contains KVStore context data. It will be written with every
// trace operation.
type TraceContext map[string]any

// Clone creates a shallow clone of a TraceContext.
func (tc TraceContext) Clone() TraceContext {
	return maps.Clone(tc)
}

// Merge merges the receiver TraceContext with the provided TraceContext argument.
func (tc TraceContext) Merge(newTc TraceContext) TraceContext {
	if tc == nil {
		tc = TraceContext{}
	}

	for k, v := range newTc {
		tc[k] = v
	}

	return tc
}
