package debug

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// captureStdout captures stdout produced during fn execution and returns it as string.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func TestRawBytes_Parses255(t *testing.T) {
	cmd := RawBytesCmd()
	cmd.SetArgs([]string{"[255]"})

	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
	})
	if out != "FF\n" {
		t.Fatalf("unexpected output: %q; want %q", out, "FF\n")
	}
}

func TestRawBytes_MixedRange(t *testing.T) {
	cmd := RawBytesCmd()
	cmd.SetArgs([]string{"[0 127 128 255]"})

	out := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute returned error: %v", err)
		}
	})
	// 0x00 0x7F 0x80 0xFF
	want := "007F80FF\n"
	if out != want {
		t.Fatalf("unexpected output: %q; want %q", out, want)
	}
}

func TestRawBytes_NegativeIsError(t *testing.T) {
	cmd := RawBytesCmd()
	// Negative should fail with ParseUint according to the fix.
	if err := cmd.RunE(cmd, []string{"[-1]"}); err == nil {
		t.Fatalf("expected error for negative input, got nil")
	}
}
