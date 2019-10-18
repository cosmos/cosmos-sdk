package cli

import (
	flag "github.com/spf13/pflag"
)

const (
	FlagSrcPort    = "src-port"
	FlagSrcChannel = "src-channel"
	FlagDenom      = "denom"
	FlagAmount     = "amount"
	FlagReceiver   = "receiver"
	FlagSource     = "source"
)

var (
	FsTransfer = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	FsTransfer.String(FlagSrcPort, "", "the source port ID")
	FsTransfer.String(FlagSrcChannel, "", "the source channel ID")
	FsTransfer.String(FlagDenom, "", "the denomination of the token to be transferred")
	FsTransfer.String(FlagAmount, "", "the amount of the token to be transferred")
	FsTransfer.String(FlagReceiver, "", "the recipient")
	FsTransfer.Bool(FlagSource, true, "indicate if the sending chain is the source chain of the token")
}
