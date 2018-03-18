package server

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"

	cmn "github.com/tendermint/tmlibs/common"
	"github.com/tendermint/tmlibs/log"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	tmtypes "github.com/tendermint/tendermint/types"
)

// InitCmd will initialize all files for tendermint,
// along with proper app_state.
// The application can pass in a function to generate
// proper state. And may want to use GenerateCoinKey
// to create default account(s).
func InitCmd(gen GenAppState, logger log.Logger) *cobra.Command {
	cmd := initCmd{
		genAppState: gen,
		logger:      logger,
	}
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis files",
		RunE:  cmd.run,
	}
}

// GenAppState can parse command-line and flag to
// generate default app_state for the genesis file.
// This is application-specific
type GenAppState func(args []string) (json.RawMessage, error)

type initCmd struct {
	genAppState GenAppState
	logger      log.Logger
}

func (c initCmd) run(cmd *cobra.Command, args []string) error {
	// Run the basic tendermint initialization,
	// set up a default genesis with no app_options
	config, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}
	err = c.initTendermintFiles(config)
	if err != nil {
		return err
	}

	// no app_options, leave like tendermint
	if c.genAppState == nil {
		return nil
	}

	// Now, we want to add the custom app_state
	appState, err := c.genAppState(args)
	if err != nil {
		return err
	}

	// And add them to the genesis file
	genFile := config.GenesisFile()
	return addGenesisState(genFile, appState)
}

// This was copied from tendermint/cmd/tendermint/commands/init.go
// so we could pass in the config and the logger.
func (c initCmd) initTendermintFiles(config *cfg.Config) error {
	// private validator
	privValFile := config.PrivValidatorFile()
	var privValidator *tmtypes.PrivValidatorFS
	if cmn.FileExists(privValFile) {
		privValidator = tmtypes.LoadPrivValidatorFS(privValFile)
		c.logger.Info("Found private validator", "path", privValFile)
	} else {
		privValidator = tmtypes.GenPrivValidatorFS(privValFile)
		privValidator.Save()
		c.logger.Info("Generated private validator", "path", privValFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if cmn.FileExists(genFile) {
		c.logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := tmtypes.GenesisDoc{
			ChainID: cmn.Fmt("test-chain-%v", cmn.RandStr(6)),
		}
		genDoc.Validators = []tmtypes.GenesisValidator{{
			PubKey: privValidator.GetPubKey(),
			Power:  10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		c.logger.Info("Generated genesis file", "path", genFile)
	}
	return nil
}

// GenesisDoc involves some tendermint-specific structures we don't
// want to parse, so we just grab it into a raw object format,
// so we can add one line.
type GenesisDoc map[string]json.RawMessage

func addGenesisState(filename string, appState json.RawMessage) error {
	bz, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var doc GenesisDoc
	err = json.Unmarshal(bz, &doc)
	if err != nil {
		return err
	}

	doc["app_state"] = appState
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, out, 0600)
}
