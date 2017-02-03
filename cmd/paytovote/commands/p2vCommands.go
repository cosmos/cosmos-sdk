package commands

import (
	"fmt"

	bcmd "github.com/tendermint/basecoin/cmd/basecoin/commands"
	"github.com/tendermint/basecoin/plugins/paytovote"
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
	P2VCreateIssueCmd = cli.Command{
		Name:  "P2VCreateIssue",
		Usage: "Create an issue which can be voted for",
		Action: func(c *cli.Context) error {
			return cmdCreateIssue(c)
		},
		Flags: []cli.Flag{
			IssueFlag,
			VoteFeeCoinFlag,
			VoteFeeAmtFlag,
		},
	}

	P2VVoteCmd = cli.Command{
		Name:  "P2VVote",
		Usage: "Vote for an existing issue",
		Action: func(c *cli.Context) error {
			return cmdVote(c)
		},
		Flags: []cli.Flag{
			IssueFlag,
			VoteForFlag,
		},
	}

	PaytovotePluginFlag = cli.BoolFlag{
		Name:  "paytovote-plugin",
		Usage: "Enable the paytovote plugin",
	}
)

func init() {
	bcmd.RegisterTxPlugin(P2VCreateIssueCmd)
	bcmd.RegisterTxPlugin(P2VVoteCmd)
	bcmd.RegisterStartPlugin(PaytovotePluginFlag,
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
	return bcmd.AppTx(c.Parent(), PaytovoteName, txBytes)
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
	return bcmd.AppTx(c.Parent(), PaytovoteName, txBytes)
}
