package keys

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"

	"github.com/cosmos/cosmos-sdk/client"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// KeyDBName is the directory under root where we store the keys
const KeyDBName = "keys"

// keybase is used to make GetKeyBase a singleton
var keybase keys.Keybase

type bechKeyOutFn func(keyInfo keys.Info) (KeyOutput, error)

// TODO make keybase take a database not load from the directory

// initialize a keybase based on the configuration
func GetKeyBase() (keys.Keybase, error) {
	rootDir := viper.GetString(cli.HomeFlag)
	return GetKeyBaseFromDir(rootDir)
}

// GetKeyInfo returns key info for a given name. An error is returned if the
// keybase cannot be retrieved or getting the info fails.
func GetKeyInfo(name string) (keys.Info, error) {
	keybase, err := GetKeyBase()
	if err != nil {
		return nil, err
	}

	return keybase.Get(name)
}

// GetPassphrase returns a passphrase for a given name. It will first retrieve
// the key info for that name if the type is local, it'll fetch input from
// STDIN. Otherwise, an empty passphrase is returned. An error is returned if
// the key info cannot be fetched or reading from STDIN fails.
func GetPassphrase(name string) (string, error) {
	var passphrase string

	keyInfo, err := GetKeyInfo(name)
	if err != nil {
		return passphrase, err
	}

	// we only need a passphrase for locally stored keys
	// TODO: (ref: #864) address security concerns
	if keyInfo.GetType() == keys.TypeLocal {
		passphrase, err = ReadPassphraseFromStdin(name)
		if err != nil {
			return passphrase, err
		}
	}

	return passphrase, nil
}

// ReadPassphraseFromStdin attempts to read a passphrase from STDIN return an
// error upon failure.
func ReadPassphraseFromStdin(name string) (string, error) {
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)

	passphrase, err := client.GetPassword(prompt, buf)
	if err != nil {
		return passphrase, fmt.Errorf("Error reading passphrase: %v", err)
	}

	return passphrase, nil
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
	Type    string `json:"type"`
	Address string `json:"address"`
	PubKey  string `json:"pub_key"`
	Seed    string `json:"seed,omitempty"`
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
	accAddr := sdk.AccAddress(info.GetPubKey().Address().Bytes())
	bechPubKey, err := sdk.Bech32ifyAccPub(info.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	return KeyOutput{
		Name:    info.GetName(),
		Type:    info.GetType().String(),
		Address: accAddr.String(),
		PubKey:  bechPubKey,
	}, nil
}

// Bech32ConsKeyOutput returns key output for a consensus node's key
// information.
func Bech32ConsKeyOutput(keyInfo keys.Info) (KeyOutput, error) {
	consAddr := sdk.ConsAddress(keyInfo.GetPubKey().Address().Bytes())

	bechPubKey, err := sdk.Bech32ifyConsPub(keyInfo.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	return KeyOutput{
		Name:    keyInfo.GetName(),
		Type:    keyInfo.GetType().String(),
		Address: consAddr.String(),
		PubKey:  bechPubKey,
	}, nil
}

// Bech32ValKeyOutput returns key output for a validator's key information.
func Bech32ValKeyOutput(keyInfo keys.Info) (KeyOutput, error) {
	valAddr := sdk.ValAddress(keyInfo.GetPubKey().Address().Bytes())

	bechPubKey, err := sdk.Bech32ifyValPub(keyInfo.GetPubKey())
	if err != nil {
		return KeyOutput{}, err
	}

	return KeyOutput{
		Name:    keyInfo.GetName(),
		Type:    keyInfo.GetType().String(),
		Address: valAddr.String(),
		PubKey:  bechPubKey,
	}, nil
}

func printKeyInfo(keyInfo keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(keyInfo)
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

func printKeyAddress(info keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.Address)
}

func printPubKey(info keys.Info, bechKeyOut bechKeyOutFn) {
	ko, err := bechKeyOut(info)
	if err != nil {
		panic(err)
	}

	fmt.Println(ko.PubKey)
}
