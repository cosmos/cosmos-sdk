package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/pkg/errors"
	isatty "github.com/tendermint/vendor-bak/github.com/mattn/go-isatty"
)

// MinPassLength is the minimum acceptable password length
const MinPassLength = 8

// if we read from non-tty, we just need to init the buffer reader once,
// in case we try to read multiple passwords (eg. update)
var buf *bufio.Reader

// GetPassword will prompt for a password one-time (to sign a tx)
// It enforces the password length
func GetPassword(prompt string) (pass string, err error) {
	if inputIsTty() {
		pass, err = speakeasy.Ask(prompt)
	} else {
		pass, err = stdinPassword()
	}
	if err != nil {
		return "", err
	}
	if len(pass) < MinPassLength {
		return "", errors.Errorf("Password must be at least %d characters", MinPassLength)
	}
	return pass, nil
}

// GetSeed will request a seed phrase from stdin and trims off
// leading/trailing spaces
func GetSeed(prompt string) (seed string, err error) {
	if inputIsTty() {
		fmt.Println(prompt)
	}
	seed, err = stdinPassword()
	seed = strings.TrimSpace(seed)
	return
}

// GetCheckPassword will prompt for a password twice to verify they
// match (for creating a new password).
// It enforces the password length. Only parses password once if
// input is piped in.
func GetCheckPassword(prompt, prompt2 string) (string, error) {
	// simple read on no-tty
	if !inputIsTty() {
		return GetPassword(prompt)
	}

	// TODO: own function???
	pass, err := GetPassword(prompt)
	if err != nil {
		return "", err
	}
	pass2, err := GetPassword(prompt2)
	if err != nil {
		return "", err
	}
	if pass != pass2 {
		return "", errors.New("Passphrases don't match")
	}
	return pass, nil
}

func inputIsTty() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

func stdinPassword() (string, error) {
	if buf == nil {
		buf = bufio.NewReader(os.Stdin)
	}
	pass, err := buf.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pass), nil
}
