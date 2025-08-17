package input

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/bgentry/speakeasy"
	isatty "github.com/mattn/go-isatty"
)

// MinPassLength is the minimum acceptable password length for security requirements
const MinPassLength = 8

// GetPassword prompts for a password with appropriate input method based on terminal type.
// For interactive terminals, uses speakeasy to hide input and provide secure prompting.
// For non-interactive input (pipes/files), reads from the provided buffer.
// Enforces minimum password length and returns the password or an error.
func GetPassword(prompt string, buf *bufio.Reader) (pass string, err error) {
	if inputIsTty() {
		pass, err = speakeasy.FAsk(os.Stderr, prompt)
	} else {
		pass, err = readLineFromBuf(buf)
	}

	if err != nil {
		return "", err
	}

	if len(pass) < MinPassLength {
		// Return the given password to the upstream client so it can handle a
		// non-STDIN failure gracefully.
		return pass, fmt.Errorf("password must be at least %d characters", MinPassLength)
	}

	return pass, nil
}

// GetConfirmation prompts for user confirmation with flexible input parsing.
// Accepts "y", "Y", "yes", "YES", and "Yes" as positive confirmations.
// Any other input (including empty) is treated as negative confirmation.
// Returns true for positive confirmation, false otherwise, with nil error.
func GetConfirmation(prompt string, r *bufio.Reader, w io.Writer) (bool, error) {
	if inputIsTty() {
		_, _ = fmt.Fprintf(w, "%s [y/N]: ", prompt)
	}

	response, err := readLineFromBuf(r)
	if err != nil {
		return false, err
	}

	response = strings.TrimSpace(response)
	if len(response) == 0 {
		return false, nil
	}

	response = strings.ToLower(response)
	if response[0] == 'y' {
		return true, nil
	}

	return false, nil
}

// GetString prompts for and reads a string input, trimming whitespace.
// For interactive terminals, displays the prompt on stderr.
// For non-interactive input, reads silently from the provided buffer.
// Returns the trimmed string or an error if reading fails.
func GetString(prompt string, buf *bufio.Reader) (string, error) {
	if inputIsTty() && prompt != "" {
		fmt.Fprintf(os.Stderr, "> %s\n", prompt)
	}

	out, err := readLineFromBuf(buf)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out), nil
}

// inputIsTty determines if stdin is connected to an interactive terminal.
// Returns true for interactive prompts where we can control input behavior
// (e.g., hide password input, request confirmation). Returns false for
// piped input from other commands where we optimize for non-interactive use.
func inputIsTty() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// readLineFromBuf reads one line from the provided buffer reader.
// Handles EOF gracefully by returning partial input if available.
// Subsequent calls reuse the same buffer, preserving input across
// multiple reads (useful for password verification scenarios).
func readLineFromBuf(buf *bufio.Reader) (string, error) {
	pass, err := buf.ReadString('\n')

	switch {
	case errors.Is(err, io.EOF):
		// If by any chance the error is EOF, but we were actually able to read
		// something from the reader then don't return the EOF error.
		// If we didn't read anything from the reader and got the EOF error, then
		// it's safe to return EOF back to the caller.
		if len(pass) > 0 {
			// exit the switch statement
			break
		}
		return "", err

	case err != nil:
		return "", err
	}

	return strings.TrimSpace(pass), nil
}
