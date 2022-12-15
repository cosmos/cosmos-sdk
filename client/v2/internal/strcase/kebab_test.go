package strcase_test

import (
	"testing"

	"cosmossdk.io/client/v2/internal/strcase"
	"gotest.tools/v3/assert"
)

func toKebab(t testing.TB) {
	cases := [][]string{
		{"testCase", "test-case"},
		{"TestCase", "test-case"},
		{"Test Case", "test-case"},
		{"TEST CASE", "test-case"},
		{"TESTCase", "test-case"},
		{"TESTCASE", "testcase"},
		{"TEST_CASE", "test-case"},
		{"Bech32", "bech32"},
		{"Bech32Address", "bech32-address"},
		{"Bech32_Address", "bech32-address"},
		{"Bech32Adress10", "bech32-adress10"},
		{"Bech32-Address10", "bech32-address10"},
		{"Bech32_Address10", "bech32-address10"},
	}
	for _, i := range cases {
		in := i[0]
		out := i[1]
		result := strcase.ToKebab(in)
		assert.Equal(t, out, result, "ToKebab(%s) = %s, want %s", in, result, out)
	}
}

func TestToKebab(t *testing.T) {
	toKebab(t)
}

func BenchmarkToKebab(b *testing.B) {
	for n := 0; n < b.N; n++ {
		toKebab(b)
	}
}
