package keys

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
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
func SignResuest(gtx *gin.Context) {
	name := gtx.Param("name")
	var m keySignBody
	body, err := ioutil.ReadAll(gtx.Request.Body)
	if err != nil {
		newError(gtx, http.StatusBadRequest, err)
		return
	}
	err = cdc.UnmarshalJSON(body, &m)
	if err != nil {
		newError(gtx, http.StatusBadRequest, err)
		return
	}
	kb, err := GetKeyBase()
	if err != nil {
		newError(gtx, http.StatusInternalServerError, err)
		return
	}

	sig, _, err := kb.Sign(name, m.Password, m.Tx)
	if err != nil {
		newError(gtx, http.StatusInternalServerError, err)
		return
	}

	encoded := base64.StdEncoding.EncodeToString(sig)

	normalResponse(gtx, []byte(encoded))
}
