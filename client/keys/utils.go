package keys

import (
	"fmt"

	"github.com/spf13/viper"

	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client"
)

var (
	// keybase is used to make GetKeyBase a singleton
	keybase keys.Keybase
)

// GetKeyBase initializes a keybase based on the configuration
func GetKeyBase() (keys.Keybase, error) {
	if keybase == nil {
		rootDir := viper.GetString(cli.HomeFlag)
		kb, err := client.GetKeyBase(rootDir)
		if err != nil {
			return nil, err
		}
		keybase = kb
	}
	return keybase, nil
}

func printInfo(info keys.Info) {
	switch viper.Get(cli.OutputFlag) {
	case "text":
		addr := info.PubKey.Address().String()
		sep := "\t\t"
		if len(info.Name) > 7 {
			sep = "\t"
		}
		fmt.Printf("%s%s%s\n", info.Name, sep, addr)
	case "json":
		json, err := MarshalJSON(info)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
	}
}

func printInfos(infos []keys.Info) {
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Println("All keys:")
		for _, i := range infos {
			printInfo(i)
		}
	case "json":
		json, err := MarshalJSON(infos)
		if err != nil {
			panic(err) // really shouldn't happen...
		}
		fmt.Println(string(json))
	}
}
