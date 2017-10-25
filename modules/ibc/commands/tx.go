package commands

import (
	"encoding/json"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/commands"
	txcmd "github.com/cosmos/cosmos-sdk/client/commands/txs"
	"github.com/cosmos/cosmos-sdk/modules/ibc"
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

// PostPacketTxCmd is CLI command to post ibc packet on the destination chain
var PostPacketTxCmd = &cobra.Command{
	Use:   "ibc-post",
	Short: "Post an ibc packet on the destination chain",
	RunE:  commands.RequireInit(postPacketTxCmd),
}

// TODO: relay!

//nolint
const (
	FlagCommit = "commit"
	FlagPacket = "packet"
)

func init() {
	fs1 := RegisterChainTxCmd.Flags()
	fs1.String(FlagCommit, "", "Filename with a commit file")

	fs2 := UpdateChainTxCmd.Flags()
	fs2.String(FlagCommit, "", "Filename with a commit file")

	fs3 := PostPacketTxCmd.Flags()
	fs3.String(FlagPacket, "", "Filename with a packet to post")
}

func registerChainTxCmd(cmd *cobra.Command, args []string) error {
	fc, err := readCommit()
	if err != nil {
		return err
	}
	tx := ibc.RegisterChainTx{fc}.Wrap()
	return txcmd.DoTx(tx)
}

func updateChainTxCmd(cmd *cobra.Command, args []string) error {
	fc, err := readCommit()
	if err != nil {
		return err
	}
	tx := ibc.UpdateChainTx{fc}.Wrap()
	return txcmd.DoTx(tx)
}

func postPacketTxCmd(cmd *cobra.Command, args []string) error {
	post, err := readPostPacket()
	if err != nil {
		return err
	}
	return txcmd.DoTx(post.Wrap())
}

func readCommit() (fc certifiers.FullCommit, err error) {
	name := viper.GetString(FlagCommit)
	if name == "" {
		return fc, errors.New("You must specify a commit file")
	}

	err = readFile(name, &fc)
	return
}

func readPostPacket() (post ibc.PostPacketTx, err error) {
	name := viper.GetString(FlagPacket)
	if name == "" {
		return post, errors.New("You must specify a packet file")
	}

	err = readFile(name, &post)
	return
}

func readFile(name string, input interface{}) (err error) {
	var f *os.File
	f, err = os.Open(name)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	// read the file as json into a seed
	j := json.NewDecoder(f)
	err = j.Decode(input)
	return errors.Wrap(err, "Invalid file")
}
