package stablejson

import (
	"encoding/base64"
	"fmt"
	"io"
)

func (opts MarshalOptions) marshalScalar(writer io.Writer, value interface{}) error {
	switch value := value.(type) {
	case string:
		_, _ = fmt.Fprintf(writer, "%q", value)
	case []byte:
		_, _ = writer.Write([]byte(`"`))
		if opts.HexBytes {
			_, _ = fmt.Fprintf(writer, "%X", value)
		} else {
			_, _ = writer.Write([]byte(base64.StdEncoding.EncodeToString(value)))
		}
		_, _ = writer.Write([]byte(`"`))
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
