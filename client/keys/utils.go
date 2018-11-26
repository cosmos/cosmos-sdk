package keys

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
)

// KeyDBName is the directory under root where we store the keys
const KeyDBName = "keys"

// keybase is used to make GetKeyBase a singleton
var keybase keys.Keybase

type bechKeyOutFn func(keyInfo keys.Info) (KeyOutput, error)

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

// TODO make keybase take a database not load from the directory

// GetKeyBase initializes a read-only KeyBase based on the configuration.
func GetKeyBase() (keys.Keybase, error) {
	rootDir := viper.GetString(cli.HomeFlag)
	return GetKeyBaseFromDir(rootDir)
}

// GetKeyBaseWithWritePerm initialize a keybase based on the configuration with write permissions.
func GetKeyBaseWithWritePerm() (keys.Keybase, error) {
	rootDir := viper.GetString(cli.HomeFlag)
	return GetKeyBaseFromDirWithWritePerm(rootDir)
}

// GetKeyBaseFromDirWithWritePerm initializes a keybase at a particular dir with write permissions.
func GetKeyBaseFromDirWithWritePerm(rootDir string) (keys.Keybase, error) {
	return getKeyBaseFromDirWithOpts(rootDir, nil)
}

// GetKeyBaseFromDir initializes a read-only keybase at a particular dir.
func GetKeyBaseFromDir(rootDir string) (keys.Keybase, error) {
	// Disabled because of the inability to create a new keys database directory
	// in the instance of when ReadOnly is set to true.
	//
	// ref: syndtr/goleveldb#240
	// return getKeyBaseFromDirWithOpts(rootDir, &opt.Options{ReadOnly: true})
	return getKeyBaseFromDirWithOpts(rootDir, nil)
}

func getKeyBaseFromDirWithOpts(rootDir string, o *opt.Options) (keys.Keybase, error) {
	if keybase == nil {
		db, err := dbm.NewGoLevelDBWithOpts(KeyDBName, filepath.Join(rootDir, "keys"), o)
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

// PostProcessResponse performs post process for rest response
func PostProcessResponse(w http.ResponseWriter, cdc *codec.Codec, response interface{}, indent bool) {
	var output []byte
	switch response.(type) {
	default:
		var err error
		if indent {
			output, err = cdc.MarshalJSONIndent(response, "", "  ")
		} else {
			output, err = cdc.MarshalJSON(response)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
	case []byte:
		output = response.([]byte)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}
