package baseapp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizeQueryMetricKey(t *testing.T) {
	testCases := []struct {
		name string
		path string
		exp  string
	}{
		{"empty", "", ""},
		{"already safe", "query_count", "query_count"},
		{"colon preserved", "store:acc", "store:acc"},
		{"store path slashes", "store/acc/key", "store_acc_key"},
		{"grpc path", "/cosmos.bank.v1beta1.Query/Balance", "_cosmos_bank_v1beta1_Query_Balance"},
		// The reported repro: a path with angle brackets produced an invalid
		// Prometheus metric name and broke /metrics. It must be neutralized.
		{"angle brackets", "/cosmos/slashing/v1beta1/signing_infos/<addr>", "_cosmos_slashing_v1beta1_signing_infos__addr_"},
		{"assorted forbidden chars", "a b=c.d-e/f<g>h{i}", "a_b_c_d_e_f_g_h_i_"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeQueryMetricKey(tc.path)
			require.Equal(t, tc.exp, got)
			// Security property: the result never contains a character outside
			// the Prometheus metric-name charset.
			for _, r := range got {
				valid := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
					(r >= '0' && r <= '9') || r == '_' || r == ':'
				require.Truef(t, valid, "unexpected character %q in sanitized key %q", r, got)
			}
		})
	}
}
