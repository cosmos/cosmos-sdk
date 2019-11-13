package tendermint

import (
	"testing"

	"github.com/stretchr/testify/require"

	yaml "gopkg.in/yaml.v2"
)

func TestString(t *testing.T) {
	dupEv := randomDuplicatedVoteEvidence()
	ev := Evidence{
		DuplicateVoteEvidence: dupEv,
		ChainID:               "mychain",
		ValidatorPower:        10,
		TotalPower:            50,
	}

	byteStr, err := yaml.Marshal(ev)
	require.Nil(t, err)
	require.Equal(t, string(byteStr), ev.String(), "Evidence String method does not work as expected")

}

func TestValidateBasic(t *testing.T) {
	dupEv := randomDuplicatedVoteEvidence()

	// good evidence
	ev := Evidence{
		DuplicateVoteEvidence: dupEv,
		ChainID:               "mychain",
		ValidatorPower:        10,
		TotalPower:            50,
	}

	err := ev.ValidateBasic()
	require.Nil(t, err, "good evidence failed on ValidateBasic: %v", err)

	// invalid duplicate evidence
	ev.DuplicateVoteEvidence.VoteA = nil
	err = ev.ValidateBasic()
	require.NotNil(t, err, "invalid duplicate evidence passed on ValidateBasic")

	// reset duplicate evidence to be valid, and set empty chainID
	dupEv = randomDuplicatedVoteEvidence()
	ev.DuplicateVoteEvidence = dupEv
	ev.ChainID = ""
	err = ev.ValidateBasic()
	require.NotNil(t, err, "invalid chain-id passed on ValidateBasic")

	// reset chainID and set 0 validator power
	ev.ChainID = "mychain"
	ev.ValidatorPower = 0
	err = ev.ValidateBasic()
	require.NotNil(t, err, "invalid validator power passed on ValidateBasic")

	// reset validator power and set invalid total power
	ev.ValidatorPower = 10
	ev.TotalPower = 9
	err = ev.ValidateBasic()
	require.NotNil(t, err, "invalid total power passed on ValidateBasic")
}
