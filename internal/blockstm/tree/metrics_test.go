package tree

import (
	"testing"

	"github.com/test-go/testify/require"
)

func TestTreeInstrumentStartBatchGetOrDefault(t *testing.T) {
	previous := treeInst
	t.Cleanup(func() {
		treeInst = previous
	})

	instrument := &treeInstrument{}
	require.NoError(t, instrument.Start(nil))
	require.NotNil(t, instrument.BatchGetOrDefault)
}
