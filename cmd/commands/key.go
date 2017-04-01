package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/spf13/cobra"

	cmn "github.com/tendermint/go-common"
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
		Run:   newKeyCmd,
	}
)

func newKeyCmd(cmd *cobra.Command, args []string) {
	key := genKey()
	keyJSON, err := json.MarshalIndent(key, "", "\t")
	fmt.Println(&key)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}
	fmt.Println(string(keyJSON))
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
	Address Address         `json:"address"`
	PubKey  crypto.PubKeyS  `json:"pub_key"`
	PrivKey crypto.PrivKeyS `json:"priv_key"`
}

// Implements Signer
func (k *Key) Sign(msg []byte) crypto.Signature {
	return k.PrivKey.Sign(msg)
}

// Generates a new validator with private key.
func genKey() *Key {
	privKey := crypto.GenPrivKeyEd25519()
	addrBytes := privKey.PubKey().Address()
	var addr Address
	copy(addr[:], addrBytes)
	return &Key{
		Address: addr,
		PubKey:  crypto.PubKeyS{privKey.PubKey()},
		PrivKey: crypto.PrivKeyS{privKey},
	}
}

func LoadKey(keyFile string) *Key {
	filePath := path.Join(BasecoinRoot(""), keyFile)
	keyJSONBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	key := new(Key)
	err = json.Unmarshal(keyJSONBytes, key)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error reading key from %v: %v\n", filePath, err))
	}

	return key
}
