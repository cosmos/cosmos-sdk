package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVoteArray(t *testing.T) {
	va := NewVoteArray(100)
	for i := 0; i < 100; i++ {
		vote := va.Get(i)
		require.False(t, vote.Missed())
		require.True(t, vote.Voted())

		vote.Miss()

		vote = va.Get(i)
		require.True(t, vote.Missed())
		require.False(t, vote.Voted())
	}

}
