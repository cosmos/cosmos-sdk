package keys

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/client"
)

// KeyDBName is the directory under root where we store the keys
const KeyDBName = "keys"

var (
	// keybase is used to make GetKeyBase a singleton
	keybase keys.Keybase
)

// used for outputting keys.Info over REST
type KeyOutput struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	// TODO add pubkey?
	// Pubkey  string `json:"pubkey"`
}

// GetKeyBase initializes a keybase based on the configuration
func GetKeyBase() (keys.Keybase, error) {
	if keybase == nil {
		rootDir := viper.GetString(cli.HomeFlag)
		db, err := dbm.NewGoLevelDB(KeyDBName, filepath.Join(rootDir, "keys"))
		if err != nil {
			return nil, err
		}
		keybase = client.GetKeyBase(db)
	}
	return keybase, nil
}

// used to set the keybase manually in test
func SetKeyBase(kb keys.Keybase) {
	keybase = kb
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
