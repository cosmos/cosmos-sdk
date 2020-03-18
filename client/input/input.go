package input

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bgentry/speakeasy"
	isatty "github.com/mattn/go-isatty"
)

// MinPassLength is the minimum acceptable password length
const MinPassLength = 8

// GetPassword will prompt for a password one-time (to sign a tx)
// It enforces the password length
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

// GetCheckPassword will prompt for a password twice to verify they
// match (for creating a new password).
// It enforces the password length. Only parses password once if
// input is piped in.
func GetCheckPassword(prompt, prompt2 string, buf *bufio.Reader) (string, error) {
	// simple read on no-tty
	if !inputIsTty() {
		return GetPassword(prompt, buf)
	}

	// TODO: own function???
	pass, err := GetPassword(prompt, buf)
	if err != nil {
		return "", err
	}
	pass2, err := GetPassword(prompt2, buf)
	if err != nil {
		return "", err
	}
	if pass != pass2 {
		return "", errors.New("passphrases don't match")
	}
	return pass, nil
}

// GetConfirmation will request user give the confirmation from stdin.
// "y", "Y", "yes", "YES", and "Yes" all count as confirmations.
// If the input is not recognized, it returns false and a nil error.
func GetConfirmation(prompt string, buf *bufio.Reader) (bool, error) {
	if inputIsTty() {
		fmt.Printf("%s [y/N]: ", prompt)
	}

	response, err := readLineFromBuf(buf)
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

// GetString simply returns the trimmed string output of a given reader.
func GetString(prompt string, buf *bufio.Reader) (string, error) {
	if inputIsTty() && prompt != "" {
		PrintPrefixed(prompt)
	}

	out, err := readLineFromBuf(buf)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// inputIsTty returns true iff we have an interactive prompt,
// where we can disable echo and request to repeat the password.
// If false, we can optimize for piped input from another command
func inputIsTty() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

// readLineFromBuf reads one line from stdin.
// Subsequent calls reuse the same buffer, so we don't lose
// any input when reading a password twice (to verify)
func readLineFromBuf(buf *bufio.Reader) (string, error) {
	pass, err := buf.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pass), nil
}

// PrintPrefixed prints a string with > prefixed for use in prompts.
func PrintPrefixed(msg string) {
	msg = fmt.Sprintf("> %s\n", msg)
	fmt.Fprint(os.Stderr, msg)
}
