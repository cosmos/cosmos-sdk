package commands

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/basecoin/client/commands"
	txcmd "github.com/tendermint/basecoin/client/commands/txs"
	"github.com/tendermint/basecoin/modules/ibc"
	"github.com/tendermint/light-client/certifiers"
)

// RegisterChainTxCmd is CLI command to register a new chain for ibc
var RegisterChainTxCmd = &cobra.Command{
	Use:   "ibc-register",
	Short: "Register a new chain",
	RunE:  commands.RequireInit(registerChainTxCmd),
}

// UpdateChainTxCmd is CLI command to update the header for an ibc chain
var UpdateChainTxCmd = &cobra.Command{
	Use:   "ibc-update",
	Short: "Add new header to an existing chain",
	RunE:  commands.RequireInit(updateChainTxCmd),
}

// TODO: post packet (query and all that jazz)

// TODO: relay!

//nolint
const (
	FlagSeed = "seed"
)

func init() {
	fs1 := RegisterChainTxCmd.Flags()
	fs1.String(FlagSeed, "", "Filename with a seed file")

	fs2 := UpdateChainTxCmd.Flags()
	fs2.String(FlagSeed, "", "Filename with a seed file")
}

func registerChainTxCmd(cmd *cobra.Command, args []string) error {
	seed, err := readSeed()
	if err != nil {
		return err
	}
	tx := ibc.RegisterChainTx{seed}.Wrap()
	return txcmd.DoTx(tx)
}

func updateChainTxCmd(cmd *cobra.Command, args []string) error {
	seed, err := readSeed()
	if err != nil {
		return err
	}
	tx := ibc.UpdateChainTx{seed}.Wrap()
	return txcmd.DoTx(tx)
}

func readSeed() (seed certifiers.Seed, err error) {
	name := viper.GetString(FlagSeed)
	if name == "" {
		return seed, errors.New("You must specify a seed file")
	}

	var f *os.File
	f, err = os.Open(name)
	if err != nil {
		return seed, errors.Wrap(err, "Cannot read seed file")
	}
	defer f.Close()

	// read the file as json into a seed
	j := json.NewDecoder(f)
	err = j.Decode(&seed)
	err = errors.Wrap(err, "Invalid seed file")
	return
}

// func readCreateRoleTxFlags() (tx basecoin.Tx, err error) {
// 	role, err := parseRole(viper.GetString(FlagRole))
// 	if err != nil {
// 		return tx, err
// 	}

// 	sigs := viper.GetInt(FlagMinSigs)
// 	if sigs < 1 {
// 		return tx, errors.Errorf("--%s must be at least 1", FlagMinSigs)
// 	}

// 	signers, err := commands.ParseActors(viper.GetString(FlagMembers))
// 	if err != nil {
// 		return tx, err
// 	}
// 	if len(signers) == 0 {
// 		return tx, errors.New("must specify at least one member")
// 	}

// 	tx = roles.NewCreateRoleTx(role, uint32(sigs), signers)
// 	return tx, nil
// }
