package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	cmttypes "github.com/cometbft/cometbft/types"

	sdkmath "cosmossdk.io/math"
	banktypes "cosmossdk.io/x/bank/types"
	stakingtypes "cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// genesisStateWithValSet returns a new genesis state with the validator set
func genesisStateWithValSet(
	codec codec.Codec,
	genesisState map[string]json.RawMessage,
	valSet *cmttypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) (map[string]json.RawMessage, error) {
	if len(genAccs) == 0 {
		return nil, errors.New("no genesis accounts provided")
	}
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey: %w", err)
		}

		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, fmt.Errorf("failed to create new any: %w", err)
		}

		validator := stakingtypes.Validator{
			OperatorAddress: sdk.ValAddress(val.Address).String(),
			ConsensusPubkey: pkAny,
			Jailed:          false,
			Status:          stakingtypes.Bonded,
			Tokens:          bondAmt,
			DelegatorShares: sdkmath.LegacyOneDec(),
			Description:     stakingtypes.Description{},
			UnbondingHeight: int64(0),
			UnbondingTime:   time.Unix(0, 0).UTC(),
			Commission: stakingtypes.NewCommission(
				sdkmath.LegacyZeroDec(),
				sdkmath.LegacyZeroDec(),
				sdkmath.LegacyZeroDec(),
			),
			MinSelfDelegation: sdkmath.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(
			delegations,
			stakingtypes.NewDelegation(
				genAccs[0].GetAddress().String(),
				sdk.ValAddress(val.Address).String(),
				sdkmath.LegacyOneDec(),
			),
		)

	}

	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(
		stakingtypes.DefaultParams(),
		validators,
		delegations,
	)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(
		stakingGenesis,
	)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(
			sdk.NewCoin(sdk.DefaultBondDenom, bondAmt),
		)
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).
			String(),
		Coins: sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	)
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	return genesisState, nil
}

// CreateRandomValidatorSet creates a validator set with one random validator
func CreateRandomValidatorSet() (*cmttypes.ValidatorSet, error) {
	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get pub key: %w", err)
	}

	// create validator set with single validator
	validator := cmttypes.NewValidator(pubKey, 1)

	return cmttypes.NewValidatorSet([]*cmttypes.Validator{validator}), nil
}

type GenesisAccount struct {
	authtypes.GenesisAccount
	Coins sdk.Coins
}

type genesisTxCodec struct {
	tx.ConfigOptions
}

func NewGenesisTxCodec(txConfigOptions tx.ConfigOptions) *genesisTxCodec {
	return &genesisTxCodec{
		txConfigOptions,
	}
}

// Decode implements transaction.Codec.
func (t *genesisTxCodec) Decode(bz []byte) (stateMachineTx, error) {
	var out stateMachineTx
	tx, err := t.ProtoDecoder(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(stateMachineTx)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}

// DecodeJSON implements transaction.Codec.
func (t *genesisTxCodec) DecodeJSON(bz []byte) (stateMachineTx, error) {
	var out stateMachineTx
	tx, err := t.JSONDecoder(bz)
	if err != nil {
		return out, err
	}

	var ok bool
	out, ok = tx.(stateMachineTx)
	if !ok {
		return out, errors.New("unexpected Tx type")
	}

	return out, nil
}
