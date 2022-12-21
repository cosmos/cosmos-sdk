package collections

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMultiIndex(t *testing.T) {
	sk, ctx := deps()
	schema := NewSchema(sk)

	mi := NewMultiIndex(schema, NewPrefix(1), "multi_index", StringKey, Uint64Key, func(value company) (string, error) {
		return value.City, nil
	})

	// we crete two reference keys for primary key 1 and 2 associated with "milan"
	require.NoError(t, mi.Reference(ctx, 1, company{City: "milan"}, nil))
	require.NoError(t, mi.Reference(ctx, 2, company{City: "milan"}, nil))

	//

}
