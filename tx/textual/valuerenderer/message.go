package valuerenderer

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type messageValueRenderer struct {
	tr *Textual
}

func NewMessageValueRenderer(t *Textual) ValueRenderer {
	return &messageValueRenderer{tr: t}
}

func (mr *messageValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	fields := v.Message().Descriptor().Fields()
	fds := make([]protoreflect.FieldDescriptor, 0, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		fds = append(fds, fields.Get(i))
	}
	sort.Slice(fds, func(i, j int) bool { return fds[i].Number() < fds[j].Number() })

	screens := make([]Screen, 1)
	screens[0].Text = fmt.Sprintf("%s object", v.Message().Descriptor().Name())

	for _, fd := range fds {
		vr, err := mr.tr.GetValueRenderer(fd)
		if err != nil {
			return nil, err
		}
		// Skip default values.
		if !v.Message().Has(fd) {
			continue
		}

		subscreens, err := vr.Format(ctx, v.Message().Get(fd))
		if err != nil {
			return nil, err
		}
		if len(subscreens) == 0 {
			return nil, fmt.Errorf("empty rendering for field %s", fd.Name())
		}

		headerScreen := Screen{
			Text:   fmt.Sprintf("%s: %s", formatFieldName(string(fd.Name())), subscreens[0].Text),
			Indent: subscreens[0].Indent + 1,
			Expert: subscreens[0].Expert,
		}
		screens = append(screens, headerScreen)

		for i := 1; i < len(subscreens); i++ {
			extraScreen := Screen{
				Text:   subscreens[i].Text,
				Indent: subscreens[i].Indent + 1,
				Expert: subscreens[i].Expert,
			}
			screens = append(screens, extraScreen)
		}
	}

	return screens, nil
}

// formatFieldName formats a field name in sentence case, as specified in:
// https://github.com/cosmos/cosmos-sdk/blob/b6f867d0b674d62e56b27aa4d00f5b6042ebac9e/docs/architecture/adr-050-sign-mode-textual-annex1.md?plain=1#L110
func formatFieldName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToTitle(name[0:1]) + strings.ReplaceAll(name[1:], "_", " ")
}

func (mr *messageValueRenderer) Parse(_ context.Context, screens []Screen) (protoreflect.Value, error) {
	panic("implement me")
}
