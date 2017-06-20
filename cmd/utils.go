package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	data "github.com/tendermint/go-wire/data"
	"github.com/tendermint/tmlibs/cli"

	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/cryptostore"
	"github.com/tendermint/go-crypto/keys/storage/filestorage"
)

const MinPassLength = 10

// GetKeyManager initializes a key manager based on the configuration
func GetKeyManager() keys.Manager {
	if manager == nil {
		// store the keys directory
		rootDir := viper.GetString(cli.HomeFlag)
		keyDir := filepath.Join(rootDir, KeySubdir)

		// TODO: smarter loading??? with language and fallback?
		codec := keys.MustLoadCodec("english")

		// and construct the key manager
		manager = cryptostore.New(
			cryptostore.SecretBox,
			filestorage.New(keyDir),
			codec,
		)
	}
	return manager
}

// if we read from non-tty, we just need to init the buffer reader once,
// in case we try to read multiple passwords (eg. update)
var buf *bufio.Reader

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

func getPassword(prompt string) (pass string, err error) {
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

func getSeed(prompt string) (seed string, err error) {
	seed, err = stdinPassword()
	seed = strings.TrimSpace(seed)
	return
}

func getCheckPassword(prompt, prompt2 string) (string, error) {
	// simple read on no-tty
	if !inputIsTty() {
		return getPassword(prompt)
	}

	// TODO: own function???
	pass, err := getPassword(prompt)
	if err != nil {
		return "", err
	}
	pass2, err := getPassword(prompt2)
	if err != nil {
		return "", err
	}
	if pass != pass2 {
		return "", errors.New("Passphrases don't match")
	}
	return pass, nil
}

func printInfo(info keys.Info) {
	switch viper.Get(cli.OutputFlag) {
	case "text":
		addr, err := data.ToText(info.Address)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		sep := "\t\t"
		if len(info.Name) > 7 {
			sep = "\t"
		}
		fmt.Printf("%s%s%s\n", info.Name, sep, addr)
	case "json":
		json, err := data.ToJSON(info)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
	}
}

func printInfos(infos keys.Infos) {
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Println("All keys:")
		for _, i := range infos {
			printInfo(i)
		}
	case "json":
		json, err := data.ToJSON(infos)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
	}
}
