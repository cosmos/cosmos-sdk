package valuerenderer

import (
	"strings"
	"testing"
)

func TestFormatDecimalNonDigits(t *testing.T) {
	badCases := []string{
		"10.a",
		"1a.10",
		"p1a10.",
		"0.10p",
		"--10",
		"12.😎😎",
		"11111111111133333333333333333333333333333a",
		"11111111111133333333333333333333333333333 192892",
	}

	for _, value := range badCases {
		value := value
		t.Run(value, func(t *testing.T) {
			s, err := formatDecimal(value)
			if err == nil {
				t.Fatal("Expected an error")
			}
			if g, w := err.Error(), "non-digits"; !strings.Contains(g, w) {
				t.Errorf("Error mismatch\nGot:  %q\nWant substring: %q", g, w)
			}
			if s != "" {
				t.Fatalf("Got a non-empty string: %q", s)
			}
		})
	}
}
