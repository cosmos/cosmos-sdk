package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseCLIProposal(t *testing.T) {
	a := assert.New(t)

	data := []byte(`{
			"group_policy_address": "cosmos15r295x4994egvckteam9skazy9kvfvzpak4naf",
			"messages": [
			  {
				"@type": "/cosmos.bank.v1beta1.MsgSend",
				"from_address": "cosmos15r295x4994egvckteam9skazy9kvfvzpak4naf",
				"to_address": "cosmos15r295x4994egvckteam9skazy9kvfvzpak4naf",
				"amount":[{"denom": "stake","amount": "10"}]
			  }
			],
			"metadata": "4pIMOgIGx1vZGU=",
			"proposers": ["cosmos15r295x4994egvckteam9skazy9kvfvzpak4naf"]
		}`)

	result, err := parseCLIProposal(data)
	a.NoError(err)
	a.Equal(result.GroupPolicyAddress, "cosmos15r295x4994egvckteam9skazy9kvfvzpak4naf")
	a.NotEmpty(result.Metadata)
	a.Equal(result.Metadata, "4pIMOgIGx1vZGU=")
	a.Equal(result.Proposers, []string{"cosmos15r295x4994egvckteam9skazy9kvfvzpak4naf"})
}
