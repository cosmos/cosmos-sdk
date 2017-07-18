package txs

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bgentry/speakeasy"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	keycmd "github.com/tendermint/go-crypto/cmd"
	"github.com/tendermint/go-crypto/keys"
	lc "github.com/tendermint/light-client"

	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/client/commands"
	"github.com/tendermint/basecoin/modules/auth"
)

// Validatable represents anything that can be Validated
type Validatable interface {
	ValidateBasic() error
}

// GetSigner returns the pub key that will sign the tx
// returns empty key if no name provided
func GetSigner() crypto.PubKey {
	name := viper.GetString(NameFlag)
	manager := keycmd.GetKeyManager()
	info, _ := manager.Get(name) // error -> empty pubkey
	return info.PubKey
}

// GetSignerAct returns the address of the signer of the tx
// (as we still only support single sig)
func GetSignerAct() (res basecoin.Actor) {
	// this could be much cooler with multisig...
	signer := GetSigner()
	if !signer.Empty() {
		res = auth.SigPerm(signer.Address())
	}
	return res
}

// Sign if it is Signable, otherwise, just convert it to bytes
func Sign(tx interface{}) (packet []byte, err error) {
	name := viper.GetString(NameFlag)
	manager := keycmd.GetKeyManager()

	if sign, ok := tx.(keys.Signable); ok {
		if name == "" {
			return nil, errors.New("--name is required to sign tx")
		}
		packet, err = signTx(manager, sign, name)
	} else if val, ok := tx.(lc.Value); ok {
		packet = val.Bytes()
	} else {
		err = errors.Errorf("Reader returned invalid tx type: %#v\n", tx)
	}
	return
}

// SignAndPostTx does all work once we construct a proper struct
// it validates the data, signs if needed, transforms to bytes,
// and posts to the node.
func SignAndPostTx(tx Validatable) (*ctypes.ResultBroadcastTxCommit, error) {
	// validate tx client-side
	err := tx.ValidateBasic()
	if err != nil {
		return nil, err
	}

	// sign the tx if needed
	packet, err := Sign(tx)
	if err != nil {
		return nil, err
	}

	// post the bytes
	node := commands.GetNode()
	return node.BroadcastTxCommit(packet)
}

// LoadJSON will read a json file from disk if --input is passed in
// template is a pointer to a struct that can hold the expected data (&MyTx{})
//
// If not data is provided, returns (false, nil)
// If data is provided and passes, returns (true, nil)
// If data is provided but not parsable, returns (true, err)
func LoadJSON(template interface{}) (bool, error) {
	input := viper.GetString(InputFlag)
	if input == "" {
		return false, nil
	}

	// load the input
	raw, err := readInput(input)
	if err != nil {
		return true, err
	}

	// parse the input
	err = json.Unmarshal(raw, template)
	if err != nil {
		return true, err
	}
	return true, nil
}

// OutputTx prints the tx result to stdout
// TODO: something other than raw json?
func OutputTx(res *ctypes.ResultBroadcastTxCommit) error {
	js, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(js))
	return nil
}

func signTx(manager keys.Manager, tx keys.Signable, name string) ([]byte, error) {
	prompt := fmt.Sprintf("Please enter passphrase for %s: ", name)
	pass, err := getPassword(prompt)
	if err != nil {
		return nil, err
	}
	err = manager.Sign(name, pass, tx)
	if err != nil {
		return nil, err
	}
	return tx.TxBytes()
}

func readInput(file string) ([]byte, error) {
	var reader io.Reader
	// get the input stream
	if file == "-" {
		reader = os.Stdin
	} else {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		reader = f
	}

	// and read it all!
	data, err := ioutil.ReadAll(reader)
	return data, errors.WithStack(err)
}

// if we read from non-tty, we just need to init the buffer reader once,
// in case we try to read multiple passwords
var buf *bufio.Reader

func inputIsTty() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

func stdinPassword() (string, error) {
	if buf == nil {
		buf = bufio.NewReader(os.Stdin)
	}
	pass, err := buf.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pass), nil
}

func getPassword(prompt string) (pass string, err error) {
	if inputIsTty() {
		pass, err = speakeasy.Ask(prompt)
	} else {
		pass, err = stdinPassword()
	}
	return
}
