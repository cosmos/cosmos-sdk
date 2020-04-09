package tx

import (
	"bufio"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/cobra"
)

type TxCmdContext struct {
	Marshaler        codec.Marshaler
	AccountRetriever AccountRetriever
	TxGenerator      Generator
}

func NewCLIContextFromTxCmd(ctx TxCmdContext, cmd *cobra.Command) context.CLIContext {
	inBuf := bufio.NewReader(cmd.InOrStdin())
	cliCtx := context.NewCLIContextWithInput(inBuf).WithMarshaler(ctx.Marshaler)
}
