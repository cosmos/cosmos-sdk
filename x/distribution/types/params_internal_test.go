package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
)

func Test_validateAuxFuncs(t *testing.T) {
	tests := []struct {
		name    string
		tax     math.LegacyDec
		wantErr bool
	}{
		{"empty math.LegacyDec", math.LegacyDec{}, true},
		{"negative", math.LegacyNewDec(-1), true},
		{"one dec", math.LegacyNewDec(1), false},
		{"two dec", math.LegacyNewDec(2), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantErr, validateCommunityTax(tt.tax) != nil)
		})
	}
}
