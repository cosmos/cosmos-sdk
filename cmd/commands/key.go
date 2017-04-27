package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	//"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/tendermint/go-crypto"
)

//commands
var (
	KeyCmd = &cobra.Command{
		Use:   "key",
		Short: "Manage keys",
	}

	NewKeyCmd = &cobra.Command{
		Use:   "new",
		Short: "Create a new private key",
		RunE:  newKeyCmd,
	}
)

func newKeyCmd(cmd *cobra.Command, args []string) error {
	key := genKey()
	keyJSON, err := json.MarshalIndent(key, "", "\t")
	if err != nil {
		return err
	}
	fmt.Println(string(keyJSON))
	return nil
}

func init() {
	//register commands
	KeyCmd.AddCommand(NewKeyCmd)
}

//---------------------------------------------
// simple implementation of a key

type Address [20]byte

func (a Address) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%x"`, a[:])), nil
}

func (a *Address) UnmarshalJSON(addrHex []byte) error {
	addr, err := hex.DecodeString(strings.Trim(string(addrHex), `"`))
	if err != nil {
		return err
	}
	copy(a[:], addr)
	return nil
}

type Key struct {
	Address Address        `json:"address"`
	PubKey  crypto.PubKey  `json:"pub_key"`
	PrivKey crypto.PrivKey `json:"priv_key"`
}

// Implements Signer
func (k *Key) Sign(msg []byte) crypto.Signature {
	return k.PrivKey.Sign(msg)
}

// Generates a new validator with private key.
func genKey() *Key {
	privKey := crypto.GenPrivKeyEd25519()
	pubKey := privKey.PubKey()
	addrBytes := pubKey.Address()
	var addr Address
	copy(addr[:], addrBytes)
	return &Key{
		Address: addr,
		PubKey:  pubKey,
		PrivKey: privKey.Wrap(),
	}
}

func LoadKey(keyFile string) (*Key, error) {
	filePath := path.Join(BasecoinRoot(""), keyFile)
	keyJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	key := new(Key)
	err = json.Unmarshal(keyJSONBytes, key)
	if err != nil {
		return nil, fmt.Errorf("Error reading key from %v: %v\n", filePath, err) //never stack trace
	}

	return key, nil
}
