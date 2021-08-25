package valuerenderer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
)


func TestFormatInt(t *testing.T) {
	
	billionStr := "1000000000"
	x, ok := types.NewIntFromString(billionStr)
	require.True(t, ok)

	d := valuerenderer.DefaultValueRenderer{}
	s, err := d.Format(x)
	require.NoError(t, err)
	require.Equal(t, s, "1,000,000,000")
}

/*
func TestFormatInt(t *testing.T) {
	v := uint64(1000000000)
	x := types.NewIntFromUint64(v)
	x64 := x.Int64()
	p := message.NewPrinter(language.English)
	s := p.Sprintf("%d", x64)
	require.Equal(t, s, "1,000,000,000")
}
*/