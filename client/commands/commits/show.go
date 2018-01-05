package commits

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/lite"
	"github.com/tendermint/tendermint/lite/files"

	"github.com/cosmos/cosmos-sdk/client/commands"
)

const (
	heightFlag = "height"
	hashFlag   = "hash"
	fileFlag   = "file"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the details of one selected commit",
	Long: `Shows the most recent downloaded key by default.
If desired, you can select by height, validator hash, or a file.
`,
	RunE:         commands.RequireInit(showCommit),
	SilenceUsage: true,
}

func init() {
	showCmd.Flags().Int(heightFlag, 0, "Show the commit with closest height to this")
	showCmd.Flags().String(hashFlag, "", "Show the commit matching the validator hash")
	showCmd.Flags().String(fileFlag, "", "Show the commit stored in the given file")
	RootCmd.AddCommand(showCmd)
}

func loadCommit(p lite.Provider, h int64, hash, file string) (fc lite.FullCommit, err error) {
	// load the commit from the proper place
	if h != 0 {
		fc, err = p.GetByHeight(h)
	} else if hash != "" {
		var vhash []byte
		vhash, err = hex.DecodeString(hash)
		if err == nil {
			fc, err = p.GetByHash(vhash)
		}
	} else if file != "" {
		fc, err = files.LoadFullCommitJSON(file)
	} else {
		// default is latest commit
		fc, err = p.LatestCommit()
	}
	return
}

func showCommit(cmd *cobra.Command, args []string) error {
	trust, _ := commands.GetProviders()

	h := int64(viper.GetInt(heightFlag))
	hash := viper.GetString(hashFlag)
	file := viper.GetString(fileFlag)
	fc, err := loadCommit(trust, h, hash, file)
	if err != nil {
		return err
	}

	// now render it!
	data, err := json.MarshalIndent(fc, "", "  ")
	fmt.Println(string(data))
	return err
}
