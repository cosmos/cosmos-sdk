package keys

import (
	"encoding/json"
	"net/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"encoding/base64"
	"fmt"
	"github.com/gorilla/mux"
)

const (
	flagFrom = "from"
	flagPassword = "password"
	flagTx = "tx"
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
	Tx            []byte `json:"tx_bytes"`
	Password      string `json:"password"`
}

// SignResuest is the handler of creating seed in swagger rest server
func SignResuest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]
	var m keySignBody

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	kb, err := GetKeyBase()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	sig, _, err := kb.Sign(name, m.Password, m.Tx)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}

	encoded := base64.StdEncoding.EncodeToString(sig)

	w.Write([]byte(encoded))
}