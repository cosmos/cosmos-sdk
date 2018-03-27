package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tendermint/go-wire/data"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/types"
	"github.com/tendermint/tmlibs/log"
)

// ShowValidator - ported from Tendermint, show this node's validator info
func ShowValidatorCmd(logger log.Logger) *cobra.Command {
	cmd := showValidator{logger}
	return &cobra.Command{
		Use:   "show_validator",
		Short: "Show this node's validator info",
		RunE:  cmd.run,
	}
}

type showValidator struct {
	logger log.Logger
}

func (s showValidator) run(cmd *cobra.Command, args []string) error {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}
	privValidator := types.LoadOrGenPrivValidatorFS(cfg.PrivValidatorFile())
	pubKeyJSONBytes, _ := data.ToJSON(privValidator.PubKey)
	fmt.Println(string(pubKeyJSONBytes))
	return nil
}
