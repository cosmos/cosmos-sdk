package stablejson

import (
	"encoding/base64"
	"fmt"
	"strings"
)

func (opts MarshalOptions) marshalScalar(writer *strings.Builder, value interface{}) error {
	switch value := value.(type) {
	case string:
		_, _ = fmt.Fprintf(writer, "%q", value)
	case []byte:
		writer.WriteString(`"`)
		if opts.HexBytes {
			_, _ = fmt.Fprintf(writer, "%X", value)
		} else {
			writer.WriteString(base64.StdEncoding.EncodeToString(value))
		}
		writer.WriteString(`"`)
	case bool:
		_, _ = fmt.Fprintf(writer, "%t", value)
	case int32:
		_, _ = fmt.Fprintf(writer, "%d", value)
	case uint32:
		_, _ = fmt.Fprintf(writer, "%d", value)
	case int64:
		_, _ = fmt.Fprintf(writer, `"%d"`, value) // quoted
	case uint64:
		_, _ = fmt.Fprintf(writer, `"%d"`, value) // quoted
	case float32:
		marshalFloat(writer, float64(value))
	case float64:
		marshalFloat(writer, value)
	default:
		return fmt.Errorf("unexpected type %T", value)
	}
	return nil
}
