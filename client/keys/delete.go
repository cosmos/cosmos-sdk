package keys

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"

	"github.com/spf13/cobra"
	"io/ioutil"
)

func deleteKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete the given key",
		RunE:  runDeleteCmd,
		Args:  cobra.ExactArgs(1),
	}
	return cmd
}

func runDeleteCmd(cmd *cobra.Command, args []string) error {
	name := args[0]

	kb, err := GetKeyBase()
	if err != nil {
		return err
	}

	_, err = kb.Get(name)
	if err != nil {
		return err
	}

	buf := client.BufferStdin()
	oldpass, err := client.GetPassword(
		"DANGER - enter password to permanently delete key:", buf)
	if err != nil {
		return err
	}

	err = kb.Delete(name, oldpass)
	if err != nil {
		return err
	}
	fmt.Println("Password deleted forever (uh oh!)")
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

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	err = cdc.UnmarshalJSON(body, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	kb, err = GetKeyBase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// TODO handle error if key is not available or pass is wrong
	err = kb.Delete(name, m.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
}
