package textual

import (
	"io"

	"cosmossdk.io/tx/textual/internal/cbor"
)

var (
	textKey   = cbor.NewUint(1)
	indentKey = cbor.NewUint(2)
	expertKey = cbor.NewUint(3)
)

// encode encodes an array of screens according to the CDDL:
//
//	screens = [* screen]
//	screen = {
//	  ? text_key: tstr,
//	  ? indent_key: uint,
//	  ? expert_key: bool,
//	}
//	text_key = 1
//	indent_key = 2
//	expert_key = 3
//
// with empty values ("", 0, false) omitted from the screen map.
func encode(screens []Screen, w io.Writer) error {
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
		// #nosec G701
		// Since we've excluded negatives, int widening is safe.
		m = m.Add(indentKey, cbor.NewUint(uint64(s.Indent)))
	}
	if s.Expert {
		m = m.Add(expertKey, cbor.NewBool(s.Expert))
	}
	return m
}
