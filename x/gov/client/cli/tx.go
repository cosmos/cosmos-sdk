package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/pkg/errors"
)

const (
	flagProposalID   = "proposalID"
	flagTitle        = "title"
	flagDescription  = "description"
	flagProposalType = "type"
	flagDeposit      = "deposit"
	flagProposer     = "proposer"
	flagDepositer    = "depositer"
	flagVoter        = "voter"
	flagOption       = "option"
)

// submit a proposal tx
func GetCmdSubmitProposal(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submitproposal",
		Short: "Submit a proposal along with an initial deposit",
		RunE: func(cmd *cobra.Command, args []string) error {
			title := viper.GetString(flagTitle)
			description := viper.GetString(flagDescription)
			proposalType := viper.GetString(flagProposalType)
			initialDeposit := viper.GetString(flagDeposit)

			// get the from address from the name flag
			from, err := sdk.GetAccAddressBech32(viper.GetString(flagProposer))
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoins(initialDeposit)
			if err != nil {
				return err
			}

			// create the message
			msg := gov.NewMsgSubmitProposal(title, description, proposalType, from, amount)
			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block:%d. Hash:%s.Response:%+v \n", res.Height, res.Hash.String(), res.DeliverTx)
			return nil
		},
	}

	cmd.Flags().String(flagTitle, "", "title of proposal")
	cmd.Flags().String(flagDescription, "", "description of proposal")
	cmd.Flags().String(flagProposalType, "", "proposalType of proposal")
	cmd.Flags().String(flagDeposit, "", "deposit of proposal")
	cmd.Flags().String(flagProposer, "", "proposer of proposal")

	return cmd
}

// set a new Deposit transaction
func GetCmdDeposit(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "deposit your token [steak] for activing proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get the from address from the name flag
			depositer, err := sdk.GetAccAddressBech32(viper.GetString(flagDepositer))
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseInt(viper.GetString(flagProposalID), 10, 64)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoins(viper.GetString(flagDeposit))
			if err != nil {
				return err
			}

			// create the message
			msg := gov.NewMsgDeposit(depositer, proposalID, amount)
			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().String(flagProposalID, "", "proposalID of proposal depositing on")
	cmd.Flags().String(flagDepositer, "", "depositer of deposit")
	cmd.Flags().String(flagDeposit, "", "amount of deposit")

	return cmd
}

// set a new Vote transaction
func GetCmdVote(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "vote for current actived proposal,option:Yes/NO/NoWithVeto/Abstain",
		RunE: func(cmd *cobra.Command, args []string) error {

			voter, err := sdk.GetAccAddressBech32(viper.GetString(flagVoter))
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseInt(viper.GetString(flagProposalID), 10, 64)
			if err != nil {
				return err
			}

			option := viper.GetString(flagOption)
			// create the message
			msg := gov.NewMsgVote(voter, proposalID, option)

			bechAddr, _ := sdk.Bech32ifyAcc(msg.Voter)

			fmt.Printf("Vote[Voter:%s,ProposalID:%d,Option:%s]", bechAddr, msg.ProposalID, msg.Option)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().String(flagProposalID, "", "proposalID of proposal voting on")
	cmd.Flags().String(flagVoter, "", "bech32 voter address")
	cmd.Flags().String(flagOption, "", "vote option {Yes, No, NoWithVeto, Abstain}")

	return cmd
}

// Command to Get a Proposal Information
func GetCmdQueryProposal(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queryproposal",
		Short: "query proposal details",
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID, err := strconv.ParseInt(viper.GetString(flagProposalID), 10, 64)
			if err != nil {
				return err
			}
			ctx := context.NewCoreContextFromViper()

			key := []byte(fmt.Sprintf("%d", proposalID) + ":proposal")
			res, err := ctx.Query(key, storeName)
			if len(res) == 0 || err != nil {
				return errors.Errorf("proposalID [%d] is not existed", proposalID)
			}

			proposal := new(gov.Proposal)
			cdc.MustUnmarshalBinary(res, proposal)
			output, err := wire.MarshalJSONIndent(cdc, proposal)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().String(flagProposalID, "", "proposalID of proposal being queried")

	return cmd
}

// Command to Get a Proposal Information
func GetCmdQueryVote(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queryvote",
		Short: "query vote",
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID, err := strconv.ParseInt(viper.GetString(flagProposalID), 10, 64)
			if err != nil {
				return err
			}

			voterAddr, err := sdk.GetAccAddressBech32(viper.GetString(flagVoter))
			if err != nil {
				return err
			}

			ctx := context.NewCoreContextFromViper()

			key := []byte(fmt.Sprintf("%d", proposalID) + ":votes:" + fmt.Sprintf("%s", voterAddr))
			res, err := ctx.Query(key, storeName)
			if len(res) == 0 || err != nil {
				return errors.Errorf("proposalID [%d] does not exist", proposalID)
			}

			vote := new(gov.Vote)
			cdc.MustUnmarshalBinary(res, vote)
			output, err := wire.MarshalJSONIndent(cdc, vote)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	cmd.Flags().String(flagProposalID, "", "proposalID of proposal voting on")
	cmd.Flags().String(flagVoter, "", "bech32 voter address")

	return cmd
}
