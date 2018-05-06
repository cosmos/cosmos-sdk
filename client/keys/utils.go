package keys

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"

	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/client"
)

// KeyDBName is the directory under root where we store the keys
const KeyDBName = "keys"

// keybase is used to make GetKeyBase a singleton
var keybase keys.Keybase

// initialize a keybase based on the configuration
func GetKeyBase() (keys.Keybase, error) {
	rootDir := viper.GetString(cli.HomeFlag)
	return GetKeyBaseFromDir(rootDir)
}

// initialize a keybase based on the configuration
func GetKeyBaseFromDir(rootDir string) (keys.Keybase, error) {
	if keybase == nil {
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

// used for outputting keys.Info over REST
type KeyOutput struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	PubKey  string `json:"pub_key"`
}

func NewKeyOutput(info keys.Info) KeyOutput {
	return KeyOutput{
		Name:    info.Name,
		Address: info.PubKey.Address().String(),
		PubKey:  strings.ToUpper(hex.EncodeToString(info.PubKey.Bytes())),
	}
}

func NewKeyOutputs(infos []keys.Info) []KeyOutput {
	kos := make([]KeyOutput, len(infos))
	for i, info := range infos {
		kos[i] = NewKeyOutput(info)
	}
	return kos
}

func printInfo(info keys.Info) {
	ko := NewKeyOutput(info)
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Printf("NAME:\tADDRESS:\t\t\t\t\tPUBKEY:\n")
		fmt.Printf("%s\t%s\t%s\n", ko.Name, ko.Address, ko.PubKey)
	case "json":
		out, err := json.MarshalIndent(ko, "", "\t")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	}
}

func printInfos(infos []keys.Info) {
	kos := NewKeyOutputs(infos)
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Printf("NAME:\tADDRESS:\t\t\t\t\tPUBKEY:\n")
		for _, ko := range kos {
			fmt.Printf("%s\t%s\t%s\n", ko.Name, ko.Address, ko.PubKey)
		}
	case "json":
		out, err := json.MarshalIndent(kos, "", "\t")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	}
}
