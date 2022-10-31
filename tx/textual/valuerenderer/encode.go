package valuerenderer

import (
	"io"

	"cosmossdk.io/tx/textual/cbor"
)

var (
	textKey   = cbor.NewUint(1)
	indentKey = cbor.NewUint(2)
	expertKey = cbor.NewUint(3)
)

// Encode encodes an array of screens according to the CDDL
func Encode(screens []Screen, w io.Writer) error {
	arr := cbor.NewArray()
	for _, s := range screens {
		arr = arr.Append(s.Cbor())
	}
	return arr.Encode(w)
}

func (s Screen) Cbor() cbor.Cbor {
	m := cbor.NewMap()
	if s.Text != "" {
		m = m.Add(textKey, cbor.NewText(s.Text))
	}
	if s.Indent > 0 {
		m = m.Add(indentKey, cbor.NewUint(uint64(s.Indent)))
	}
	if s.Expert {
		m = m.Add(expertKey, cbor.NewBool(s.Expert))
	}
	return m
}
