package commands

import (
	"testing"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/spf13/viper"
)

func TestVoteCmd(t *testing.T) {
	cdc := app.MakeCodec()
	cmd := VoteCmd(cdc)

	viper.Set(flagvoter,"BD37661BE7F88C52E217A707DE1416311B141FDC")
	viper.Set(flagproposalID,0)
	viper.Set(flagoption,"Yes")

	viper.Set("name","gov")



	cmd.Execute()
}
