package keys

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	keys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KeyDBName is the directory under root where we store the keys
const KeyDBName = "keys"

// keybase is used to make GetKeyBase a singleton
var keybase keys.Keybase

// TODO make keybase take a database not load from the directory

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
	Name    string         `json:"name"`
	Type    string         `json:"type"`
	Address sdk.AccAddress `json:"address"`
	PubKey  string         `json:"pub_key"`
	Seed    string         `json:"seed,omitempty"`
}

// create a list of KeyOutput in bech32 format
func Bech32KeysOutput(infos []keys.Info) ([]KeyOutput, error) {
	kos := make([]KeyOutput, len(infos))
	for i, info := range infos {
		ko, err := Bech32KeyOutput(info)
		if err != nil {
			return nil, err
		}
		kos[i] = ko
	}
	return kos, nil
}

// create a KeyOutput in bech32 format
func Bech32KeyOutput(info keys.Info) (KeyOutput, error) {
	account := sdk.AccAddress(info.GetPubKey().Address().Bytes())
	bechPubKey, err := sdk.Bech32ifyAccPub(info.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}
	return KeyOutput{
		Name:    info.GetName(),
		Type:    info.GetType(),
		Address: account,
		PubKey:  bechPubKey,
	}, nil
}

func printInfo(info keys.Info) {
	ko, err := Bech32KeyOutput(info)
	if err != nil {
		panic(err)
	}
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Printf("NAME:\tTYPE:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
		printKeyOutput(ko)
	case "json":
		out, err := MarshalJSON(ko)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	}
}

func printInfos(infos []keys.Info) {
	kos, err := Bech32KeysOutput(infos)
	if err != nil {
		panic(err)
	}
	switch viper.Get(cli.OutputFlag) {
	case "text":
		fmt.Printf("NAME:\tTYPE:\tADDRESS:\t\t\t\t\t\tPUBKEY:\n")
		for _, ko := range kos {
			printKeyOutput(ko)
		}
	case "json":
		out, err := MarshalJSON(kos)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(out))
	}
}

func printKeyOutput(ko KeyOutput) {
	fmt.Printf("%s\t%s\t%s\t%s\n", ko.Name, ko.Type, ko.Address, ko.PubKey)
}
