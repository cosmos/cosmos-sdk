package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/pkg/errors"
)

const (
	flagProposalID        = "proposal-id"
	flagTitle             = "title"
	flagDescription       = "description"
	flagProposalType      = "type"
	flagDeposit           = "deposit"
	flagVoter             = "voter"
	flagOption            = "option"
	flagDepositer         = "depositer"
	flagStatus            = "status"
	flagLatestProposalIDs = "latest"
)

// submit a proposal tx
func GetCmdSubmitProposal(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-proposal",
		Short: "Submit a proposal along with an initial deposit",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			title := viper.GetString(flagTitle)
			description := viper.GetString(flagDescription)
			strProposalType := viper.GetString(flagProposalType)
			initialDeposit := viper.GetString(flagDeposit)

			// get the from address from the name flag
			fromAddr, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoins(initialDeposit)
			if err != nil {
				return err
			}

			proposalType, err := gov.ProposalTypeFromString(strProposalType)
			if err != nil {
				return err
			}

			// create the message
			msg := gov.NewMsgSubmitProposal(title, description, proposalType, fromAddr, amount)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			// proposalID must be returned, and it is a part of response
			ctx.PrintResponse = true
			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String(flagTitle, "", "title of proposal")
	cmd.Flags().String(flagDescription, "", "description of proposal")
	cmd.Flags().String(flagProposalType, "", "proposalType of proposal")
	cmd.Flags().String(flagDeposit, "", "deposit of proposal")

	return cmd
}

// set a new Deposit transaction
func GetCmdDeposit(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "deposit tokens for activing proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			// get the from address from the name flag
			depositerAddr, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			proposalID := viper.GetInt64(flagProposalID)

			amount, err := sdk.ParseCoins(viper.GetString(flagDeposit))
			if err != nil {
				return err
			}

			// create the message
			msg := gov.NewMsgDeposit(depositerAddr, proposalID, amount)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			// build and sign the transaction, then broadcast to Tendermint
			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String(flagProposalID, "", "proposalID of proposal depositing on")
	cmd.Flags().String(flagDeposit, "", "amount of deposit")

	return cmd
}

// set a new Vote transaction
func GetCmdVote(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "vote for an active proposal, options: Yes/No/NoWithVeto/Abstain",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithDecoder(authcmd.GetAccountDecoder(cdc))

			voterAddr, err := ctx.GetFromAddress()
			if err != nil {
				return err
			}

			proposalID := viper.GetInt64(flagProposalID)

			option := viper.GetString(flagOption)

			byteVoteOption, err := gov.VoteOptionFromString(option)
			if err != nil {
				return err
			}

			// create the message
			msg := gov.NewMsgVote(voterAddr, proposalID, byteVoteOption)

			err = msg.ValidateBasic()
			if err != nil {
				return err
			}

			fmt.Printf("Vote[Voter:%s,ProposalID:%d,Option:%s]",
				voterAddr.String(), msg.ProposalID, msg.Option.String())

			// build and sign the transaction, then broadcast to Tendermint
			err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().String(flagProposalID, "", "proposalID of proposal voting on")
	cmd.Flags().String(flagOption, "", "vote option {Yes, No, NoWithVeto, Abstain}")

	return cmd
}

// Command to Get a Proposal Information
func GetCmdQueryProposal(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-proposal",
		Short: "query proposal details",
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID := viper.GetInt64(flagProposalID)

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.QueryStore(gov.KeyProposal(proposalID), storeName)
			if len(res) == 0 || err != nil {
				return errors.Errorf("proposalID [%d] is not existed", proposalID)
			}

			var proposal gov.Proposal
			cdc.MustUnmarshalBinary(res, &proposal)
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

// nolint: gocyclo
// Command to Query Proposals
func GetCmdQueryProposals(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-proposals",
		Short: "query proposals with optional filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			bechDepositerAddr := viper.GetString(flagDepositer)
			bechVoterAddr := viper.GetString(flagVoter)
			strProposalStatus := viper.GetString(flagStatus)
			latestProposalsIDs := viper.GetInt64(flagLatestProposalIDs)

			var err error
			var voterAddr sdk.AccAddress
			var depositerAddr sdk.AccAddress
			var proposalStatus gov.ProposalStatus

			if len(bechDepositerAddr) != 0 {
				depositerAddr, err = sdk.AccAddressFromBech32(bechDepositerAddr)
				if err != nil {
					return err
				}
			}

			if len(bechVoterAddr) != 0 {
				voterAddr, err = sdk.AccAddressFromBech32(bechVoterAddr)
				if err != nil {
					return err
				}
			}

			if len(strProposalStatus) != 0 {
				proposalStatus, err = gov.ProposalStatusFromString(strProposalStatus)
				if err != nil {
					return err
				}
			}

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.QueryStore(gov.KeyNextProposalID, storeName)
			if err != nil {
				return err
			}
			var maxProposalID int64
			cdc.MustUnmarshalBinary(res, &maxProposalID)

			matchingProposals := []gov.Proposal{}

			if latestProposalsIDs == 0 {
				latestProposalsIDs = maxProposalID
			}

			for proposalID := maxProposalID - latestProposalsIDs; proposalID < maxProposalID; proposalID++ {
				if voterAddr != nil {
					res, err = ctx.QueryStore(gov.KeyVote(proposalID, voterAddr), storeName)
					if err != nil || len(res) == 0 {
						continue
					}
				}

				if depositerAddr != nil {
					res, err = ctx.QueryStore(gov.KeyDeposit(proposalID, depositerAddr), storeName)
					if err != nil || len(res) == 0 {
						continue
					}
				}

				res, err = ctx.QueryStore(gov.KeyProposal(proposalID), storeName)
				if err != nil || len(res) == 0 {
					continue
				}

				var proposal gov.Proposal
				cdc.MustUnmarshalBinary(res, &proposal)

				if len(strProposalStatus) != 0 {
					if proposal.GetStatus() != proposalStatus {
						continue
					}
				}

				matchingProposals = append(matchingProposals, proposal)
			}

			if len(matchingProposals) == 0 {
				fmt.Println("No matching proposals found")
				return nil
			}

			for _, proposal := range matchingProposals {
				fmt.Printf("  %d - %s\n", proposal.GetProposalID(), proposal.GetTitle())
			}
			return nil
		},
	}

	cmd.Flags().String(flagLatestProposalIDs, "", "(optional) limit to latest [number] proposals. Defaults to all proposals")
	cmd.Flags().String(flagDepositer, "", "(optional) filter by proposals deposited on by depositer")
	cmd.Flags().String(flagVoter, "", "(optional) filter by proposals voted on by voted")
	cmd.Flags().String(flagStatus, "", "(optional) filter proposals by proposal status")

	return cmd
}

// Command to Get a Proposal Information
func GetCmdQueryVote(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-vote",
		Short: "query vote",
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID := viper.GetInt64(flagProposalID)

			voterAddr, err := sdk.AccAddressFromBech32(viper.GetString(flagVoter))
			if err != nil {
				return err
			}

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.QueryStore(gov.KeyVote(proposalID, voterAddr), storeName)
			if len(res) == 0 || err != nil {
				return errors.Errorf("proposalID [%d] does not exist", proposalID)
			}

			var vote gov.Vote
			cdc.MustUnmarshalBinary(res, &vote)
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

// Command to Get a Proposal Information
func GetCmdQueryVotes(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query-votes",
		Short: "query votes on a proposal",
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID := viper.GetInt64(flagProposalID)

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.QueryStore(gov.KeyProposal(proposalID), storeName)
			if len(res) == 0 || err != nil {
				return errors.Errorf("proposalID [%d] does not exist", proposalID)
			}

			var proposal gov.Proposal
			cdc.MustUnmarshalBinary(res, &proposal)

			if proposal.GetStatus() != gov.StatusVotingPeriod {
				fmt.Println("Proposal not in voting period.")
				return nil
			}

			res2, err := ctx.QuerySubspace(cdc, gov.KeyVotesSubspace(proposalID), storeName)
			if err != nil {
				return err
			}

			var votes []gov.Vote
			for i := 0; i < len(res2); i++ {
				var vote gov.Vote
				cdc.MustUnmarshalBinary(res2[i].Value, &vote)
				votes = append(votes, vote)
			}

			output, err := wire.MarshalJSONIndent(cdc, votes)
			if err != nil {
				return err
			}

			fmt.Println(string(output))

			return nil
		},
	}

	cmd.Flags().String(flagProposalID, "", "proposalID of which proposal's votes are being queried")

	return cmd
}
