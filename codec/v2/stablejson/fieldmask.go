package stablejson

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	pathsName protoreflect.Name = "paths"
)

func marshalFieldMask(writer *strings.Builder, value protoreflect.Message) error {
	field := value.Descriptor().Fields().ByName(pathsName)
	if field == nil {
		return fmt.Errorf("expected to find field %s", pathsName)
	}

	paths := value.Get(field).List()
	n := paths.Len()
	strs := make([]string, n)
	for i := 0; i < n; i++ {
		strs[i] = paths.Get(i).String()
	}
	_, _ = fmt.Fprintf(writer, "%q", strings.Join(strs, ","))
	return nil
}
