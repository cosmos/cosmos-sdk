package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/tendermint/lite"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	wire "github.com/cosmos/cosmos-sdk/wire"

	authcmd "github.com/cosmos/cosmos-sdk/x/auth/commands"
	ibcm "github.com/cosmos/cosmos-sdk/x/ibc"
)

type openCommander struct {
	cdc       *wire.Codec
	parser    sdk.AccountDecoder
	mainStore string
	ibcStore  string
}

func IBCOpenCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := openCommander{
		cdc:       cdc,
		parser:    authcmd.GetAccountDecoder(cdc),
		ibcStore:  "ibc",
		mainStore: "main",
	}

	cmd := &cobra.Command{
		Use: "open",
		Run: cmdr.runIBCOpen,
	}

	cmd.Flags().String(FlagFromChainNode, "tcp://localhost:46657", "<host>:<port> to tendermint rpc interface for this chain")
	cmd.Flags().String(FlagFromChainID, "", "Chain ID for ibc node to open channel to")

	cmd.MarkFlagRequired(FlagFromChainNode)
	cmd.MarkFlagRequired(FlagFromChainID)
	cmd.MarkFlagRequired(client.FlagChainID)

	viper.BindPFlag(FlagFromChainID, cmd.Flags().Lookup(FlagFromChainID))

	return cmd
}

func (c openCommander) runIBCOpen(cmd *cobra.Command, args []string) {
	fromChainID := viper.GetString(FlagFromChainID)

	fromCtx := context.NewCoreContextFromViper().WithNodeURI(viper.GetString(FlagFromChainNode))
	toCtx := context.NewCoreContextFromViper().WithNodeURI(viper.GetString(client.FlagChainID))
	address, err := toCtx.GetFromAddress()
	if err != nil {
		panic(err)
	}

	node, err := fromCtx.GetNode()
	if err != nil {
		panic(err)
	}
	gen := int64(1)
	commit, err := node.Commit(&gen)
	if err != nil {
		panic(err)
	}
	valset, err := node.Validators(&gen)
	if err != nil {
		panic(err)
	}

	msg := ibcm.OpenChannelMsg{
		SrcChain: fromChainID,
		ROT: lite.NewFullCommit(
			lite.Commit(commit.SignedHeader),
			tmtypes.NewValidatorSet(valset.Validators),
		),
		Signer: address,
	}

	name := viper.GetString(client.FlagName)
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)
	passphrase, err := client.GetPassword(prompt, buf)
	if err != nil {
		panic(err)
	}

	tx, err := toCtx.SignAndBuild(name, passphrase, msg, c.cdc)
	if err != nil {
		panic(err)
	}
	res, err := toCtx.BroadcastTx(tx)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
}
