package keys

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/viper"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"

	"github.com/spf13/cobra"
)

const (
	flagYes   = "yes"
	flagForce = "force"
)

func deleteKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete the given key",
		Long: `Delete a key from the store.

Note that removing offline or ledger keys will remove
only the public key references stored locally, i.e.
private keys stored in a ledger device cannot be deleted with
gaiacli.
`,
		RunE: runDeleteCmd,
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().BoolP(flagYes, "y", false,
		"Skip confirmation prompt when deleting offline or ledger key references")
	cmd.Flags().BoolP(flagForce, "f", false,
		"Remove the key unconditionally without asking for the passphrase")
	return cmd
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	name := args[0]

	kb, err := NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}

	info, err := kb.Get(name)
	if err != nil {
		return err
	}

	buf := client.BufferStdin()
	if info.GetType() == keys.TypeLedger || info.GetType() == keys.TypeOffline {
		if !viper.GetBool(flagYes) {
			if err := confirmDeletion(buf); err != nil {
				return err
			}
		}
		if err := kb.Delete(name, "", true); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "Public key reference deleted")
		return nil
	}

	// skip passphrase check if run with --force
	skipPass := viper.GetBool(flagForce)
	var oldpass string
	if !skipPass {
		if oldpass, err = client.GetPassword(
			"DANGER - enter password to permanently delete key:", buf); err != nil {
			return err
		}
	}

	err = kb.Delete(name, oldpass, skipPass)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stderr, "Key deleted forever (uh oh!)")
	return nil
}

func confirmDeletion(buf *bufio.Reader) error {
	answer, err := client.GetConfirmation("Key reference will be deleted. Continue?", buf)
	if err != nil {
		return err
	}
	if !answer {
		return errors.New("aborted")
	}
	return nil
}

////////////////////////
// REST

// delete key request REST body
type DeleteKeyBody struct {
	Password string `json:"password"`
}

// delete key REST handler
func DeleteKeyRequestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var kb keys.Keybase
	var m DeleteKeyBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	kb, err = NewKeyBaseFromHomeFlag()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	err = kb.Delete(name, m.Password, false)
	if keyerror.IsErrKeyNotFound(err) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(err.Error()))
		return
	} else if keyerror.IsErrWrongPassword(err) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(err.Error()))
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
