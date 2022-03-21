package valuerenderer

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ValueRenderer defines an interface to produce formated output for Int,Dec,Coin types as well as parse a string to Coin or Uint.
type ValueRenderer interface {
	Format(context.Context, interface{}) ([]string, error)
	Parse(context.Context, []string) (interface{}, error)
}

type DefaultValueRenderer struct{}

var _ ValueRenderer = DefaultValueRenderer{}

func (r DefaultValueRenderer) Format(ctx context.Context, v interface{}) ([]string, error) {
	switch v := v.(type) {
	case int:
		return r.formatNumber(ctx, strconv.Itoa(v))

	default:
		return nil, fmt.Errorf("value renderers cannot format value %s of type %T", v, v)
	}
}

func (r DefaultValueRenderer) Parse(context.Context, []string) (interface{}, error) {
	panic("TODO")
}

func (r DefaultValueRenderer) formatNumber(_ context.Context, v string) ([]string, error) {
	// Remove leading and trailing zeroes.
	v = strings.Trim(v, "0")
	// Add `'` delimiter every 3 integral digits
	startOffset := 3
	for outputIndex := len(v); outputIndex > startOffset; {
		outputIndex -= 3
		v = v[:outputIndex] + "," + v[outputIndex:]
	}

	return []string{v}, nil
}
