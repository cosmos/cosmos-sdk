package keys

import (
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
)

// CMD

// listKeysCmd represents the list command
var listKeysCmd = &cobra.Command{
	Use:   "list",
	Short: "List all keys",
	Long: `Return a list of all public keys stored by this key manager
along with their associated name and address.`,
	RunE: runListCmd,
}

func pseudoListing(kb keys.Keybase) (infos []keys.Info, err error) {
	infos, err = kb.List()
	if err != nil {
		return nil, err
	}
	// Pseudo-item for Ledger
	path := []uint32{44, 60, 0, 0, 0} // TODO
	ledger, lerr := crypto.NewPrivKeyLedgerSecp256k1(path)
	if lerr == nil {
		ledgerInfo := keys.Info{
			Name:         "ledger",
			PubKey:       ledger.PubKey(),
			PrivKeyArmor: "",
		}
		infos = append(infos, ledgerInfo)
	}
	return infos, err
}

func runListCmd(cmd *cobra.Command, args []string) error {
	kb, err := GetKeyBase()
	if err != nil {
		return err
	}

	infos, err := pseudoListing(kb)

	if err == nil {
		printInfos(infos)
	}
	return err
}

/////////////////////////
// REST

// query key list REST handler
func QueryKeysRequestHandler(w http.ResponseWriter, r *http.Request) {
	kb, err := GetKeyBase()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	infos, err := pseudoListing(kb)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	// an empty list will be JSONized as null, but we want to keep the empty list
	if len(infos) == 0 {
		w.Write([]byte("[]"))
		return
	}
	keysOutput := make([]KeyOutput, len(infos))
	for i, info := range infos {
		keysOutput[i] = KeyOutput{Name: info.Name, Address: info.PubKey.Address().String()}
	}
	output, err := json.MarshalIndent(keysOutput, "", "  ")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	w.Write(output)
}
