package valuerenderer

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

type messageValueRenderer struct {
	tr *Textual
}

var _ ValueRenderer = (*messageValueRenderer)(nil)

func (mr *messageValueRenderer) Format(ctx context.Context, v protoreflect.Value, w io.Writer) error {
	fields := v.Message().Descriptor().Fields()
	fds := make([]protoreflect.FieldDescriptor, 0, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		fds = append(fds, fields.Get(i))
	}
	sort.Slice(fds, func(i, j int) bool { return fds[i].Number() < fds[j].Number() })

	fmt.Fprintf(w, "%s object\n", v.Message().Descriptor().Name())
	buf := &bytes.Buffer{}
	for _, fd := range fds {
		vr, err := mr.tr.GetValueRenderer(fd)
		if err != nil {
			return err
		}
		// Skip default values.
		if !v.Message().Has(fd) {
			continue
		}

		buf.Reset()
		if err := vr.Format(ctx, v.Message().Get(fd), buf); err != nil {
			return fmt.Errorf("failed to format subfield %s: %w", fd.FullName(), err)
		}

		sc := bufio.NewScanner(buf)
		if sc.Scan() {
			str := sc.Text()
			fmt.Fprintf(w, "> %s: %s\n", formatFieldName(string(fd.Name())), str)
		}
		for sc.Scan() {
			str := sc.Text()
			// Only add a space after the > if the field isn't already nested.
			nesting := "> "
			if str[0] == '>' {
				nesting = ">"
			}
			fmt.Fprintf(w, "%s%s\n", nesting, str)
		}
		if err := sc.Err(); err != nil {
			return err
		}
	}

	return nil
}

// formatFieldName formats a field name in sentence case, as specified in:
// https://github.com/cosmos/cosmos-sdk/blob/b6f867d0b674d62e56b27aa4d00f5b6042ebac9e/docs/architecture/adr-050-sign-mode-textual-annex1.md?plain=1#L110
func formatFieldName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[0:1]) + strings.ReplaceAll(name[1:], "_", " ")
}

func (mr *messageValueRenderer) Parse(_ context.Context, r io.Reader) (protoreflect.Value, error) {
	panic("implement me")
}
