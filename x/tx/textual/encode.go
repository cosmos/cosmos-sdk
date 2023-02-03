package textual

import (
	"io"

	"cosmossdk.io/x/tx/textual/internal/cbor"
)

var (
	titleKey   = cbor.NewUint(1)
	contentKey = cbor.NewUint(2)
	indentKey  = cbor.NewUint(3)
	expertKey  = cbor.NewUint(4)
)

// encode encodes an array of screens according to the CDDL:
//
//	screens = [* screen]
//	screen = {
//	  ? title_key: tstr,
//	  ? content_key: tstr,
//	  ? indent_key: uint,
//	  ? expert_key: bool,
//	}
//	text_key = 1
//	content_key = 2
//	indent_key = 3
//	expert_key = 4
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
	if s.Title != "" {
		m = m.Add(titleKey, cbor.NewText(s.Title))
	}
	if s.Content != "" {
		m = m.Add(contentKey, cbor.NewText(s.Content))
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
