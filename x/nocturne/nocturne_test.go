
package nocturne

import (
	"testing"
)

func TestHelloNocturne(t *testing.T) {
	expected := "Hello from Nocturne!"
	actual := HelloNocturne()
	if actual != expected {
		t.Errorf("HelloNocturne() = %q, want %q", actual, expected)
	}
}
