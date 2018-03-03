package commands

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/builder"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/cosmos/cosmos-sdk/examples/basecoin/x/cool"
)

// SendTxCommand will create a send tx and sign it with the given key
func WhatCoolTxCmd(cdc *wire.Codec) *cobra.Command {
	cmdr := commander{cdc}
	return &cobra.Command{
		Use:   "whatcool [answer]",
		Short: "What's cooler than being cool?",
		RunE:  cmdr.whatCoolTxCmd,
	}
}

type commander struct {
	cdc *wire.Codec
}

func (c commander) whatCoolTxCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide an answer")
	}

	// get the from address from the name flag
	from, err := builder.GetFromAddress()
	if err != nil {
		return err
	}

	// create the message
	msg := cool.NewWhatCoolMsg(from, args[0])

	// build and sign the transaction, then broadcast to Tendermint
	res, err := builder.SignBuildBroadcast(msg, c.cdc)
	if err != nil {
		return err
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
	return nil
}
