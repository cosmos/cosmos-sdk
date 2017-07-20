package seeds

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/light-client/certifiers"

	"github.com/tendermint/basecoin/client/commands"
)

const (
	heightFlag = "height"
	hashFlag   = "hash"
	fileFlag   = "file"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the details of one selected seed",
	Long: `Shows the most recent downloaded key by default.
If desired, you can select by height, validator hash, or a file.
`,
	RunE:         commands.RequireInit(showSeed),
	SilenceUsage: true,
}

func init() {
	showCmd.Flags().Int(heightFlag, 0, "Show the seed with closest height to this")
	showCmd.Flags().String(hashFlag, "", "Show the seed matching the validator hash")
	showCmd.Flags().String(fileFlag, "", "Show the seed stored in the given file")
	RootCmd.AddCommand(showCmd)
}

func loadSeed(p certifiers.Provider, h int, hash, file string) (seed certifiers.Seed, err error) {
	// load the seed from the proper place
	if h != 0 {
		seed, err = p.GetByHeight(h)
	} else if hash != "" {
		var vhash []byte
		vhash, err = hex.DecodeString(hash)
		if err == nil {
			seed, err = p.GetByHash(vhash)
		}
	} else if file != "" {
		seed, err = certifiers.LoadSeed(file)
	} else {
		// default is latest seed
		seed, err = certifiers.LatestSeed(p)
	}
	return
}

func showSeed(cmd *cobra.Command, args []string) error {
	trust, _ := commands.GetProviders()

	h := viper.GetInt(heightFlag)
	hash := viper.GetString(hashFlag)
	file := viper.GetString(fileFlag)
	seed, err := loadSeed(trust, h, hash, file)
	if err != nil {
		return err
	}

	// now render it!
	data, err := json.MarshalIndent(seed, "", "  ")
	fmt.Println(string(data))
	return err
}
