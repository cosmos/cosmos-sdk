package server

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tendermint/go-wire/data"
	"github.com/tendermint/tendermint/types"
)

// ShowValidator - ported from Tendermint, show this node's validator info
func ShowValidatorCmd(ctx *Context) *cobra.Command {
	cmd := showValidator{ctx}
	return &cobra.Command{
		Use:   "show_validator",
		Short: "Show this node's validator info",
		RunE:  cmd.run,
	}
}

type showValidator struct {
	context *Context
}

func (s showValidator) run(cmd *cobra.Command, args []string) error {
	cfg := s.context.Config
	privValidator := types.LoadOrGenPrivValidatorFS(cfg.PrivValidatorFile())
	pubKeyJSONBytes, err := data.ToJSON(privValidator.PubKey)
	if err != nil {
		return err
	}
	fmt.Println(string(pubKeyJSONBytes))
	return nil
}
