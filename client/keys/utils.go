package keys

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	keys "github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"

	"github.com/cosmos/cosmos-sdk/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
	Name    string        `json:"name"`
	Address sdk.Address   `json:"address"`
	PubKey  crypto.PubKey `json:"pub_key"`
}

func NewKeyOutput(info keys.Info) KeyOutput {
	return KeyOutput{
		Name:    info.Name,
		Address: sdk.Address(info.PubKey.Address().Bytes()),
		PubKey:  info.PubKey,
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
		fmt.Printf("NAME:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
		printKeyOutput(ko)
	case "json":
		out, err := MarshalJSON(ko)
		//out, err := json.MarshalIndent(kos, "", "\t")
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
		fmt.Printf("NAME:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
		for _, ko := range kos {
			printKeyOutput(ko)
		}
	case "json":
		out, err := MarshalJSON(kos)
		//out, err := json.MarshalIndent(kos, "", "\t")
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	}
}

func printKeyOutput(ko KeyOutput) {
	bechAccount, err := sdk.Bech32CosmosifyAcc(ko.Address)
	if err != nil {
		panic(err)
	}
	bechPubKey, err := sdk.Bech32CosmosifyAccPub(ko.PubKey)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\t%s\t%s\n", ko.Name, bechAccount, bechPubKey)
}
