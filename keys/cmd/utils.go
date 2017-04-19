package cmd

import (
	"fmt"

	"github.com/bgentry/speakeasy"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	data "github.com/tendermint/go-data"
	keys "github.com/tendermint/go-crypto/keys"
)

const PassLength = 10

func getPassword(prompt string) (string, error) {
	pass, err := speakeasy.Ask(prompt)
	if err != nil {
		return "", err
	}
	if len(pass) < PassLength {
		return "", errors.Errorf("Password must be at least %d characters", PassLength)
	}
	return pass, nil
}

func getCheckPassword(prompt, prompt2 string) (string, error) {
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
	switch viper.Get(OutputFlag) {
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
	switch viper.Get(OutputFlag) {
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
