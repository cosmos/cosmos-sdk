package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/spf13/viper"

	"github.com/tendermint/go-crypto"
	"github.com/tendermint/tmlibs/cli"
)

//---------------------------------------------
// simple implementation of a key

// Address - public address for a key
type Address [20]byte

// MarshalJSON - marshal the json bytes of the address
func (a Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%x"`, a[:])), nil
}

// UnmarshalJSON - unmarshal the json bytes of the address
func (a *Address) UnmarshalJSON(addrHex []byte) error {
	addr, err := hex.DecodeString(strings.Trim(string(addrHex), `"`))
	if err != nil {
		return err
	}
	copy(a[:], addr)
	return nil
}

// Key - full private key
type Key struct {
	Address Address        `json:"address"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	PrivKey crypto.PrivKey `json:"priv_key"`
}

// Sign - Implements Signer
func (k *Key) Sign(msg []byte) crypto.Signature {
	return k.PrivKey.Sign(msg)
}

// LoadKey - load key from json file
func LoadKey(keyFile string) (*Key, error) {
	filePath := keyFile

	if !strings.HasPrefix(keyFile, "/") && !strings.HasPrefix(keyFile, ".") {
		rootDir := viper.GetString(cli.HomeFlag)
		filePath = path.Join(rootDir, keyFile)
	}

	keyJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	key := new(Key)
	err = json.Unmarshal(keyJSONBytes, key)
	if err != nil {
		return nil, fmt.Errorf("Error reading key from %v: %v", filePath, err) //never stack trace
	}

	return key, nil
}
