package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/pflag"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func parseSignalProposalFlags(fs *pflag.FlagSet) (*types.MsgSignal, error) {
	msg := &types.MsgSignal{}
	proposalFile, _ := fs.GetString(FlagProposal)

	if proposalFile == "" {
		msg.Title, _ = fs.GetString(FlagTitle)
		msg.Description, _ = fs.GetString(FlagDescription)
		return msg, nil
	}

	for _, flag := range ProposalFlags {
		if v, _ := fs.GetString(flag); v != "" {
			return nil, fmt.Errorf("--%s flag provided alongside --proposal, which is a noop", flag)
		}
	}

	contents, err := ioutil.ReadFile(proposalFile)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(contents, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func parseMsgs(msgFile string) ([]sdk.Msg, error) {
	msgs := []sdk.Msg{}

	msgBytes, err := ioutil.ReadFile(msgFile)
	if err != nil {
		return msgs, err
	}

	err = json.Unmarshal(msgBytes, msgs)
	return msgs, err
}
