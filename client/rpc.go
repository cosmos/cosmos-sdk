package client

import (
	"context"
	"fmt"
	"strings"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

// get the current blockchain height
func GetChainHeight(clientCtx Context) (int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return -1, err
	}

	status, err := node.Status(context.Background())
	if err != nil {
		return -1, err
	}

	height := status.SyncInfo.LatestBlockHeight
	return height, nil
}

// Validator output
type ValidatorOutput struct {
	Address          sdk.ConsAddress    `json:"address"`
	PubKey           cryptotypes.PubKey `json:"pub_key"`
	ProposerPriority int64              `json:"proposer_priority"`
	VotingPower      int64              `json:"voting_power"`
}

// Validators at a certain height output in bech32 format
type ResultValidatorsOutput struct {
	BlockHeight int64             `json:"block_height"`
	Validators  []ValidatorOutput `json:"validators"`
	Total       uint64            `json:"total"`
}

func (rvo ResultValidatorsOutput) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("block height: %d\n", rvo.BlockHeight))
	b.WriteString(fmt.Sprintf("total count: %d\n", rvo.Total))

	for _, val := range rvo.Validators {
		b.WriteString(
			fmt.Sprintf(`
  Address:          %s
  Pubkey:           %s
  ProposerPriority: %d
  VotingPower:      %d
		`,
				val.Address, val.PubKey, val.ProposerPriority, val.VotingPower,
			),
		)
	}

	return b.String()
}

func validatorOutput(validator *tmtypes.Validator) (ValidatorOutput, error) {
	pk, err := cryptocodec.FromTmPubKeyInterface(validator.PubKey)
	if err != nil {
		return ValidatorOutput{}, err
	}

	return ValidatorOutput{
		Address:          sdk.ConsAddress(validator.Address),
		PubKey:           pk,
		ProposerPriority: validator.ProposerPriority,
		VotingPower:      validator.VotingPower,
	}, nil
}

// GetValidators from client
func GetValidators(ctx context.Context, clientCtx Context, height *int64, page, limit *int) (ResultValidatorsOutput, error) {
	// get the node
	node, err := clientCtx.GetNode()
	if err != nil {
		return ResultValidatorsOutput{}, err
	}

	validatorsRes, err := node.Validators(ctx, height, page, limit)
	if err != nil {
		return ResultValidatorsOutput{}, err
	}

	total := validatorsRes.Total
	if validatorsRes.Total < 0 {
		total = 0
	}
	out := ResultValidatorsOutput{
		BlockHeight: validatorsRes.BlockHeight,
		Validators:  make([]ValidatorOutput, len(validatorsRes.Validators)),
		Total:       uint64(total),
	}
	for i := 0; i < len(validatorsRes.Validators); i++ {
		out.Validators[i], err = validatorOutput(validatorsRes.Validators[i])
		if err != nil {
			return out, err
		}
	}

	return out, nil
}
