package commands

// import (
// 	"fmt"

// 	"github.com/pkg/errors"
// 	"github.com/spf13/cobra"
// 	"github.com/spf13/viper"

// 	"github.com/cosmos/cosmos-sdk/client"
// 	"github.com/cosmos/cosmos-sdk/client/builder"
// 	"github.com/cosmos/cosmos-sdk/examples/democoin/x/cool"
// 	sdk "github.com/cosmos/cosmos-sdk/types"
// 	"github.com/cosmos/cosmos-sdk/wire"
// 	"github.com/cosmos/cosmos-sdk/x/gov"
// )

// const (
// 	flagTitle          = "title"
// 	flagDescription    = "description"
// 	flagProposalType   = "type"
// 	flagInitialDeposit = "deposit"
// )

// func SubmitProposalCmd(cdc *wire.Codec) *cobra.Command {
// 	cmdr := commander{cdc}
// 	cmd := &cobra.Command{
// 		Use:   "bond",
// 		Short: "Bond to a validator",
// 		RunE:  cmdr.bondTxCmd,
// 	}
// 	cmd.Flags().String(flagStake, "", "Amount of coins to stake")
// 	cmd.Flags().String(flagValidator, "", "Validator address to stake")
// 	return cmd
// }

// // submit a proposal tx
// func SubmitProposalCmd(cdc *wire.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "submitproposal [title] [description] [proposaltype] [initialdeposit]",
// 		Short: "Submit a proposal along with an initial deposit",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			if len(args) != 1 || len(args[0]) == 0 {
// 				return errors.New("You must provide an answer")
// 			}

// 			// get the from address from the name flag
// 			from, err := builder.GetFromAddress()
// 			if err != nil {
// 				return err
// 			}

// 			// create the message
// 			msg := gov.NewSubmitProposalMsg(arg[0], arg[1], arg[2], from, arg[3])
// 			chainID := viper.GetString(client.FlagChainID)
// 			sequence := int64(viper.GetInt(client.FlagSequence))

// 			signMsg := sdk.StdSignMsg{
// 				ChainID:   chainID,
// 				Sequences: []int64{sequence},
// 				Msg:       msg,
// 			}

// 			// build and sign the transaction, then broadcast to Tendermint
// 			res, err := builder.SignBuildBroadcast(signMsg, cdc)
// 			if err != nil {
// 				return err
// 			}

// 			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
// 			return nil
// 		},
// 	}
// }

// // set a new cool trend transaction
// func DepositCmd(cdc *wire.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "setcool [answer]",
// 		Short: "You're so cool, tell us what is cool!",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			if len(args) != 1 || len(args[0]) == 0 {
// 				return errors.New("You must provide an answer")
// 			}

// 			// get the from address from the name flag
// 			from, err := builder.GetFromAddress()
// 			if err != nil {
// 				return err
// 			}

// 			// create the message
// 			msg := cool.NewSetTrendMsg(from, args[0])
// 			chainID := viper.GetString(client.FlagChainID)
// 			sequence := int64(viper.GetInt(client.FlagSequence))

// 			signMsg := sdk.StdSignMsg{
// 				ChainID:   chainID,
// 				Sequences: []int64{sequence},
// 				Msg:       msg,
// 			}

// 			// build and sign the transaction, then broadcast to Tendermint
// 			res, err := builder.SignBuildBroadcast(signMsg, cdc)
// 			if err != nil {
// 				return err
// 			}

// 			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
// 			return nil
// 		},
// 	}
// }

// // set a new cool trend transaction
// func VoteCmd(cdc *wire.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "setcool [answer]",
// 		Short: "You're so cool, tell us what is cool!",
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			if len(args) != 1 || len(args[0]) == 0 {
// 				return errors.New("You must provide an answer")
// 			}

// 			// get the from address from the name flag
// 			from, err := builder.GetFromAddress()
// 			if err != nil {
// 				return err
// 			}

// 			// create the message
// 			msg := cool.NewSetTrendMsg(from, args[0])
// 			chainID := viper.GetString(client.FlagChainID)
// 			sequence := int64(viper.GetInt(client.FlagSequence))

// 			signMsg := sdk.StdSignMsg{
// 				ChainID:   chainID,
// 				Sequences: []int64{sequence},
// 				Msg:       msg,
// 			}

// 			// build and sign the transaction, then broadcast to Tendermint
// 			res, err := builder.SignBuildBroadcast(signMsg, cdc)
// 			if err != nil {
// 				return err
// 			}

// 			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
// 			return nil
// 		},
// 	}
// }
