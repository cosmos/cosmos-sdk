package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// validator struct to define the fields of the validator
type validator struct {
	Amount              string `json:"amount"`
	From                string `json:"from"`
	PubKey              string `json:"pubkey"`
	Moniker             string `json:"moniker"`
	Identity            string `json:"identity"`
	Website             string `json:"website"`
	Securiy             string `json:"security"`
	Details             string `json:"details"`
	CommissionRate      string `json:"commission-rate"`
	CommissionMaxRate   string `json:"commission-max-rate"`
	CommissionMaxChange string `json:"commission-max-change-rate"`
	MinSelfDelegation   string `json:"min-self-delegation"`
}

func parseValidatorJSON(cdc codec.Codec, path string) (from, amount, pubkey, moniker, cm_rate, cm_max_rate, cm_max_change_rate, min_self_del string, err error) {
	var validator validator

	contents, err := os.ReadFile(path)
	if err != nil {
		return "", "", "", "", "", "", "", "", err
	}

	err = json.Unmarshal(contents, &validator)
	if err != nil {
		return "", "", "", "", "", "", "", "", err
	}

	if err := validator.validate(); err != nil {
		return "", "", "", "", "", "", "", "", err
	}

	return validator.From, validator.Amount, validator.PubKey, validator.Moniker, validator.CommissionRate, validator.CommissionMaxRate, validator.CommissionMaxChange, validator.MinSelfDelegation, nil
}

// validate checks that the required fields are present in the validator json.
func (v validator) validate() error {
	if v.Amount == "" {
		return fmt.Errorf("must specify amount of coins to bond")
	}
	if v.From == "" {
		return fmt.Errorf("must specify name or address of from key")
	}
	if v.PubKey == "" {
		return fmt.Errorf("must specify the JSON encoded pubkey")
	}
	if v.Moniker == "" {
		return fmt.Errorf("must specify the moniker name")
	}
	if v.CommissionRate == "" {
		return fmt.Errorf("must specify initial commission rate percentage")
	}
	if v.CommissionMaxRate == "" {
		return fmt.Errorf("must specify maximum commission rate percentage")
	}
	if v.CommissionMaxChange == "" {
		return fmt.Errorf("must specify maximum commission change rate percentage (per day)")
	}
	if v.MinSelfDelegation == "" {
		return fmt.Errorf("minimum self delegation must be a positive integer")
	}
	return nil
}

func buildCommissionRates(rateStr, maxRateStr, maxChangeRateStr string) (commission types.CommissionRates, err error) {
	if rateStr == "" || maxRateStr == "" || maxChangeRateStr == "" {
		return commission, errors.New("must specify all validator commission parameters")
	}

	rate, err := sdk.NewDecFromStr(rateStr)
	if err != nil {
		return commission, err
	}

	maxRate, err := sdk.NewDecFromStr(maxRateStr)
	if err != nil {
		return commission, err
	}

	maxChangeRate, err := sdk.NewDecFromStr(maxChangeRateStr)
	if err != nil {
		return commission, err
	}

	commission = types.NewCommissionRates(rate, maxRate, maxChangeRate)

	return commission, nil
}
