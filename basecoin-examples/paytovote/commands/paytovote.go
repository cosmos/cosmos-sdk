package commands

import (
	"fmt"

	"github.com/tendermint/basecoin-examples/paytovote"
	bcmd "github.com/tendermint/basecoin/cmd/commands"
	"github.com/tendermint/basecoin/types"
	"github.com/urfave/cli"
)

const PaytovoteName = "paytovote"

var (
	//common flag
	IssueFlag = cli.StringFlag{
		Name:  "issue",
		Value: "default issue",
		Usage: "name of the issue to generate or vote for",
	}

	//createIssue flags
	VoteFeeCoinFlag = cli.StringFlag{
		Name:  "voteFeeCoin",
		Value: "",
		Usage: "the fee's coin type to vote for the issue",
	}
	VoteFeeAmtFlag = cli.IntFlag{
		Name:  "voteFeeAmt",
		Value: 0,
		Usage: "the fee amount of coin type VoteCoinFlag to vote for the issue",
	}

	//vote flag
	VoteForFlag = cli.BoolFlag{
		Name:  "voteFor",
		Usage: "set to true when vote be cast is a vote-for the issue, false if vote-against",
	}
)

var (
	P2VCmd = cli.Command{
		Name:  "paytovote",
		Usage: "Send transactions to the paytovote plugin",
		Subcommands: []cli.Command{
			P2VCreateIssueCmd,
			P2VVoteCmd,
		},
	}

	P2VCreateIssueCmd = cli.Command{
		Name:  "create-issue",
		Usage: "Create an issue which can be voted for",
		Action: func(c *cli.Context) error {
			return cmdCreateIssue(c)
		},
		Flags: append(bcmd.TxFlags,
			IssueFlag,
			VoteFeeCoinFlag,
			VoteFeeAmtFlag,
		),
	}

	P2VVoteCmd = cli.Command{
		Name:  "vote",
		Usage: "Vote for an existing issue",
		Action: func(c *cli.Context) error {
			return cmdVote(c)
		},
		Flags: append(bcmd.TxFlags,
			IssueFlag,
			VoteForFlag,
		),
	}
)

func init() {
	bcmd.RegisterTxSubcommand(P2VCmd)
	bcmd.RegisterStartPlugin(PaytovoteName,
		func() types.Plugin { return paytovote.New() })
}

func cmdCreateIssue(c *cli.Context) error {
	issue := c.String(IssueFlag.Name)
	feeCoin := c.String(VoteFeeCoinFlag.Name)
	feeAmt := int64(c.Int(VoteFeeAmtFlag.Name))

	voteFee := types.Coins{{feeCoin, feeAmt}}
	createIssueFee := types.Coins{{"issueToken", 1}} //manually set the cost to create a new issue

	txBytes := paytovote.NewCreateIssueTxBytes(issue, voteFee, createIssueFee)

	fmt.Println("Issue creation transaction sent")
	return bcmd.AppTx(c, PaytovoteName, txBytes)
}

func cmdVote(c *cli.Context) error {
	issue := c.String(IssueFlag.Name)
	voteFor := c.Bool(VoteForFlag.Name)

	var voteTB byte = paytovote.TypeByteVoteFor
	if !voteFor {
		voteTB = paytovote.TypeByteVoteAgainst
	}

	txBytes := paytovote.NewVoteTxBytes(issue, voteTB)

	fmt.Println("Vote transaction sent")
	return bcmd.AppTx(c, PaytovoteName, txBytes)
}
