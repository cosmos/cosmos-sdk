package client

import (
	bam "github.com/cosmos/cosmos-sdk/baseapp"
	bank "github.com/cosmos/cosmos-sdk/x/bank"
)

// build the sendTx msg
func BuildMsg(from bam.Address, to bam.Address, coins bam.Coins) bam.Msg {
	input := bank.NewInput(from, coins)
	output := bank.NewOutput(to, coins)
	msg := bank.NewMsgSend([]bank.Input{input}, []bank.Output{output})
	return msg
}
