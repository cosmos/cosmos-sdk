package input

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type fakeReader struct {
	fnc func(p []byte) (int, error)
}

func (f fakeReader) Read(p []byte) (int, error) {
	return f.fnc(p)
}

var _ io.Reader = fakeReader{}

func TestStringWhitespace(t *testing.T) {
	for _, tc := range []struct{ name, input, raw, trimmed string }{
		{"LF", " \tcafé\u00a0 \n", " \tcafé\u00a0 ", "café"},
		{"CRLF", " \tcafé\u00a0 \r\n", " \tcafé\u00a0 ", "café"},
		{"EOF", " \tcafé\u00a0 \r", " \tcafé\u00a0 \r", "café"},
		{"whitespace only", " \n", " ", ""},
		{"empty line", "\n", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := GetStringRaw("", bufio.NewReader(strings.NewReader(tc.input)))
			require.NoError(t, err)
			require.Equal(t, tc.raw, raw)
			trimmed, err := GetString("", bufio.NewReader(strings.NewReader(tc.input)))
			require.NoError(t, err)
			require.Equal(t, tc.trimmed, trimmed)
		})
	}
	// Reading a raw passphrase must not consume the confirmation line.
	buf := bufio.NewReader(strings.NewReader(" secret \n secret \n"))
	for range 2 {
		pass, err := GetStringRaw("", buf)
		require.NoError(t, err)
		require.Equal(t, " secret ", pass)
	}
	_, err := GetStringRaw("", buf)
	require.ErrorIs(t, err, io.EOF)

	t.Run("keyring password remains trimmed", func(t *testing.T) {
		if inputIsTty() {
			t.Skip("requires non-TTY input")
		}
		pass, err := GetPassword("", bufio.NewReader(strings.NewReader(" 12345678 \n")))
		require.NoError(t, err)
		require.Equal(t, "12345678", pass)
		_, err = GetPassword("", bufio.NewReader(strings.NewReader(" 1234567 \n")))
		require.Error(t, err)
	})
}

func TestReadLineFromBuf(t *testing.T) {
	var fr fakeReader

	t.Run("it correctly returns the password when reader returns EOF", func(t *testing.T) {
		fr.fnc = func(p []byte) (int, error) {
			return copy(p, []byte("hello")), io.EOF
		}
		buf := bufio.NewReader(fr)

		pass, err := readLineFromBuf(buf)
		require.NoError(t, err)
		require.Equal(t, "hello", pass)
	})

	t.Run("it returns EOF if reader has been exhausted", func(t *testing.T) {
		fr.fnc = func(p []byte) (int, error) {
			return 0, io.EOF
		}
		buf := bufio.NewReader(fr)

		_, err := readLineFromBuf(buf)
		require.ErrorIs(t, err, io.EOF)
	})

	t.Run("it returns the error if it's not EOF regardless if it read something or not", func(t *testing.T) {
		expectedErr := errors.New("oh no")
		fr.fnc = func(p []byte) (int, error) {
			return copy(p, []byte("hello")), expectedErr
		}
		buf := bufio.NewReader(fr)

		_, err := readLineFromBuf(buf)
		require.ErrorIs(t, err, expectedErr)
	})
}
