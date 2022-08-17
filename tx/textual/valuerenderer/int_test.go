package valuerenderer

import (
	"strings"
	"testing"
)

func TestFormatIntegerNonDigits(t *testing.T) {
	badCases := []string{
		"a10",
		"1a10",
		"p1a10",
		"10p",
		"--10",
		"😎😎",
		"11111111111133333333333333333333333333333a",
		"11111111111133333333333333333333333333333 192892",
	}

	for _, value := range badCases {
		value := value
		t.Run(value, func(t *testing.T) {
			s, err := formatInteger(value)
			if err == nil {
				t.Fatal("Expected an error")
			}
			if g, w := err.Error(), "but got non-digits in"; !strings.Contains(g, w) {
				t.Errorf("Error mismatch\nGot:  %q\nWant substring: %q", g, w)
			}
			if s != "" {
				t.Fatalf("Got a non-empty string: %q", s)
			}
		})
	}
}
