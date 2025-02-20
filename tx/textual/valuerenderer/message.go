package valuerenderer

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
)

type messageValueRenderer struct {
	tr      *Textual
	msgDesc protoreflect.MessageDescriptor
	fds     []protoreflect.FieldDescriptor
}

func NewMessageValueRenderer(t *Textual, msgDesc protoreflect.MessageDescriptor) ValueRenderer {
	fields := msgDesc.Fields()
	fds := make([]protoreflect.FieldDescriptor, 0, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		fds = append(fds, fields.Get(i))
	}
	sort.Slice(fds, func(i, j int) bool { return fds[i].Number() < fds[j].Number() })

	return &messageValueRenderer{tr: t, msgDesc: msgDesc, fds: fds}
}

func (mr *messageValueRenderer) header() string {
	return fmt.Sprintf("%s object", mr.msgDesc.Name())
}

func (mr *messageValueRenderer) Format(ctx context.Context, v protoreflect.Value) ([]Screen, error) {
	fullName := v.Message().Descriptor().FullName()
	wantFullName := mr.msgDesc.FullName()
	if fullName != wantFullName {
		return nil, fmt.Errorf(`bad message type: want "%s", got "%s"`, wantFullName, fullName)
	}

	screens := make([]Screen, 1)
	screens[0].Text = mr.header()

	for _, fd := range mr.fds {
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

var nilValue = protoreflect.Value{}

func (mr *messageValueRenderer) Parse(ctx context.Context, screens []Screen) (protoreflect.Value, error) {
	if len(screens) == 0 {
		return nilValue, fmt.Errorf("expect at least one screen")
	}

	wantHeader := fmt.Sprintf("%s object", mr.msgDesc.Name())
	if screens[0].Text != wantHeader {
		return nilValue, fmt.Errorf(`bad header: want "%s", got "%s"`, wantHeader, screens[0].Text)
	}
	if screens[0].Indent != 0 {
		return nilValue, fmt.Errorf("bad message indentation: want 0, got %d", screens[0].Indent)
	}

	msgType, err := protoregistry.GlobalTypes.FindMessageByName(mr.msgDesc.FullName())
	if err != nil {
		return nilValue, err
	}
	msg := msgType.New()
	idx := 1

	for _, fd := range mr.fds {
		if idx >= len(screens) {
			// remaining fields are default
			break
		}

		vr, err := mr.tr.GetValueRenderer(fd)
		if err != nil {
			return nilValue, err
		}

		if screens[idx].Indent != 1 {
			return nilValue, fmt.Errorf("bad message indentation: want 1, got %d", screens[idx].Indent)
		}

		prefix := formatFieldName(string(fd.Name())) + ": "
		if !strings.HasPrefix(screens[idx].Text, prefix) {
			// we must have skipped this fd because of a default value
			continue
		}

		// Make a new screen without the prefix
		subscreens := make([]Screen, 1)
		subscreens[0] = screens[idx]
		subscreens[0].Text = strings.TrimPrefix(screens[idx].Text, prefix)
		subscreens[0].Indent--
		idx++

		// Gather nested screens
		for idx < len(screens) && screens[idx].Indent > 1 {
			scr := screens[idx]
			scr.Indent--
			subscreens = append(subscreens, scr)
			idx++
		}

		val, err := vr.Parse(ctx, subscreens)
		if err != nil {
			return nilValue, err
		}
		msg.Set(fd, val)
	}

	if idx > len(screens) {
		return nilValue, fmt.Errorf("leftover screens")
	}

	return protoreflect.ValueOfMessage(msg), nil
}
