package errors

import (
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeInt64FromUint64(t *testing.T) {
	cases := map[string]struct {
		gasWanted      uint64
		gasUsed        uint64
		expectedWanted int64
		expectedUsed   int64
	}{
		"normal values": {
			gasWanted:      1000,
			gasUsed:        500,
			expectedWanted: 1000,
			expectedUsed:   500,
		},
		"zero values": {
			gasWanted:      0,
			gasUsed:        0,
			expectedWanted: 0,
			expectedUsed:   0,
		},
		"max int64 values": {
			gasWanted:      math.MaxInt64,
			gasUsed:        math.MaxInt64,
			expectedWanted: math.MaxInt64,
			expectedUsed:   math.MaxInt64,
		},
		"overflow case": {
			gasWanted:      math.MaxInt64 + 1,
			gasUsed:        math.MaxInt64 + 1,
			expectedWanted: math.MaxInt64,
			expectedUsed:   math.MaxInt64,
		},
		"max uint64 values": {
			gasWanted:      math.MaxUint64,
			gasUsed:        math.MaxUint64,
			expectedWanted: math.MaxInt64,
			expectedUsed:   math.MaxInt64,
		},
	}

	for testName, tc := range cases {
		t.Run(testName, func(t *testing.T) {
			// Test ResponseCheckTxWithEvents
			response := ResponseCheckTxWithEvents(nil, tc.gasWanted, tc.gasUsed, nil, false)
			require.Equal(t, tc.expectedWanted, response.GasWanted, testName+" - CheckTx GasWanted")
			require.Equal(t, tc.expectedUsed, response.GasUsed, testName+" - CheckTx GasUsed")

			// Test ResponseExecTxResultWithEvents
			execResponse := ResponseExecTxResultWithEvents(nil, tc.gasWanted, tc.gasUsed, nil, false)
			require.Equal(t, tc.expectedWanted, execResponse.GasWanted, testName+" - ExecTx GasWanted")
			require.Equal(t, tc.expectedUsed, execResponse.GasUsed, testName+" - ExecTx GasUsed")
		})
	}
}
