package keys

import (
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"net/http"
)

const (
	flagFrom     = "from"
	flagPassword = "password"
	flagTx       = "tx"
)

func init() {
	keySignCmd.Flags().String(flagFrom, "", "Name of private key with which to sign")
	keySignCmd.Flags().String(flagPassword, "", "Password of private key")
	keySignCmd.Flags().String(flagTx, "", "Base64 encoded tx data for sign")
}

var keySignCmd = &cobra.Command{
	Use:   "sign",
	Short: "Sign user specified data",
	Long:  `Sign user data with specified key and password`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name := viper.GetString(flagFrom)
		password := viper.GetString(flagPassword)
		tx := viper.GetString(flagTx)

		decodedTx, err := base64.StdEncoding.DecodeString(tx)
		if err != nil {
			return err
		}

		kb, err := GetKeyBase()
		if err != nil {
			return err
		}

		sig, _, err := kb.Sign(name, password, decodedTx)
		if err != nil {
			return err
		}
		encoded := base64.StdEncoding.EncodeToString(sig)
		fmt.Println(string(encoded))
		return nil
	},
}

type keySignBody struct {
	Tx       []byte `json:"tx_bytes"`
	Password string `json:"password"`
}

// SignResuest is the handler of creating seed in swagger rest server
func SignResuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var m keySignBody

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

	kb, err := GetKeyBase()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	sig, _, err := kb.Sign(name, m.Password, m.Tx)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	encoded := base64.StdEncoding.EncodeToString(sig)

	w.Write([]byte(encoded))
}
