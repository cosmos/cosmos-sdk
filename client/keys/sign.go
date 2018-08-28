package keys

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"github.com/cosmos/cosmos-sdk/client/httputils"
	"net/http"
)

type SignBody struct {
	Tx            []byte `json:"tx_bytes"`
	Password      string `json:"password"`
}
/*
var showKeysCmd = &cobra.Command{
	Use:   "sign <name>",
	Short: "Show key info for the given name",
	Long:  `Return public details of one local key.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		info, err := getKey(name)
		if err != nil {
			return err
		}

		showAddress := viper.GetBool(FlagAddress)
		showPublicKey := viper.GetBool(FlagPublicKey)
		outputSet := cmd.Flag(cli.OutputFlag).Changed
		if showAddress && showPublicKey {
			return errors.New("cannot use both --address and --pubkey at once")
		}
		if outputSet && (showAddress || showPublicKey) {
			return errors.New("cannot use --output with --address or --pubkey")
		}
		if showAddress {
			printKeyAddress(info)
			return nil
		}
		if showPublicKey {
			printPubKey(info)
			return nil
		}

		printInfo(info)
		return nil
	},
}
*/
// SignResuest is the handler of creating seed in swagger rest server
func SignResuest(gtx *gin.Context) {
	name := gtx.Param("name")
	var m SignBody
	body, err := ioutil.ReadAll(gtx.Request.Body)
	if err != nil {
		httputils.NewError(gtx, http.StatusBadRequest, err)
		return
	}
	err = cdc.UnmarshalJSON(body, &m)
	if err != nil {
		httputils.NewError(gtx, http.StatusBadRequest, err)
		return
	}
	kb, err := GetKeyBase()
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}

	sig, _, err := kb.Sign(name, m.Password, m.Tx)
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}

	output, err := cdc.MarshalJSON(sig)
	if err != nil {
		httputils.NewError(gtx, http.StatusInternalServerError, err)
		return
	}

	httputils.NormalResponse(gtx, output)
}