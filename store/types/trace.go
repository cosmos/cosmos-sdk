package types

import (
	"io"
)

//----------------------------------------

// TraceContext contains TraceKVStore context data. It will be written with
// every trace operation.
type TraceContext map[string]interface{}

// Tracer is pair of io.Writer and TraceContext
type Tracer struct {
	Writer  io.Writer
	Context TraceContext
}

// Enabled returns if tracing is enabled.
func (t *Tracer) Enabled() bool {
	return t.Context != nil
}

// WithTracer sets the writer that the underlying stores of the owner
// multistore will utilize to trace operations.
func (t *Tracer) SetWriter(w io.Writer) {
	t.Writer = w
}

// WithContext sets the tracing context for a Tracer. It is implied that
// the caller should update the context when necessary between tracing
// operations.
func (t *Tracer) SetContext(tc TraceContext) {
	if t.Context != nil {
		for k, v := range tc {
			t.Context[k] = v
		}
	} else {
		t.Context = tc
	}
}

// ResetContext resets the current tracing context.
func (t *Tracer) ResetContext() {
	t.Context = nil
}
