package testnet

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GenesisBuilder enables constructing a genesis file,
// following a builder pattern.
//
// None of the methods on GenesisBuilder return an error,
// choosing instead to panic.
// GenesisBuilder is only intended for use in tests,
// where inputs are predetermined and expected to succeed.
type GenesisBuilder struct {
	amino *codec.LegacyAmino
	codec *codec.ProtoCodec

	// The value used in ChainID.
	// Some other require this value,
	// so store it as a field instead of re-parsing it from JSON.
	chainID string

	// The outer JSON object.
	// Most data goes into app_state, but there are some top-level fields.
	outer map[string]json.RawMessage

	// Many of GenesisBuilder's methods operate on the app_state JSON object,
	// so we track that separately and nest it inside outer upon a call to JSON().
	appState map[string]json.RawMessage

	gentxs []sdk.Tx
}

// NewGenesisBuilder returns an initialized GenesisBuilder.
//
// The returned GenesisBuilder has an initial height of 1
// and a genesis_time of the current time when the function was called.
func NewGenesisBuilder() *GenesisBuilder {
	ir := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(ir)
	stakingtypes.RegisterInterfaces(ir)
	banktypes.RegisterInterfaces(ir)
	authtypes.RegisterInterfaces(ir)
	pCodec := codec.NewProtoCodec(ir)

	return &GenesisBuilder{
		amino: codec.NewLegacyAmino(),
		codec: pCodec,

		outer: map[string]json.RawMessage{
			"initial_height": json.RawMessage(`"1"`),
			"genesis_time": json.RawMessage(
				strconv.AppendQuote(nil, time.Now().UTC().Format(time.RFC3339Nano)),
			),
		},
		appState: map[string]json.RawMessage{},
	}
}

// GenTx emulates the gentx CLI, creating a message to create a validator
// represented by val, with "amount" self delegation,
// and signed by privVal.
func (b *GenesisBuilder) GenTx(privVal secp256k1.PrivKey, val cmttypes.GenesisValidator, amount sdk.Coin) *GenesisBuilder {
	if b.chainID == "" {
		panic(fmt.Errorf("(*GenesisBuilder).GenTx must not be called before (*GenesisBuilder).ChainID"))
	}

	pubKey, err := cryptocodec.FromCmtPubKeyInterface(val.PubKey)
	if err != nil {
		panic(err)
	}

	// Produce the create validator message.
	msg, err := stakingtypes.NewMsgCreateValidator(
		privVal.PubKey().Address().Bytes(),
		pubKey,
		amount,
		stakingtypes.Description{
			Moniker: "TODO",
		},
		stakingtypes.CommissionRates{
			Rate:          sdk.MustNewDecFromStr("0.1"),
			MaxRate:       sdk.MustNewDecFromStr("0.2"),
			MaxChangeRate: sdk.MustNewDecFromStr("0.01"),
		},
		sdk.OneInt(),
	)
	if err != nil {
		panic(err)
	}
	valAddr, err := sdk.ValAddressFromBech32(msg.ValidatorAddress)
	if err != nil {
		panic(err)
	}

	msg.DelegatorAddress = sdk.AccAddress(valAddr).String()

	if err := msg.ValidateBasic(); err != nil {
		panic(err)
	}

	txConf := authtx.NewTxConfig(b.codec, tx.DefaultSignModes)

	txb := txConf.NewTxBuilder()
	if err := txb.SetMsgs(msg); err != nil {
		panic(err)
	}

	const signMode = signing.SignMode_SIGN_MODE_DIRECT

	// Need to set the signature object on the tx builder first,
	// otherwise we end up signing a different total message
	// compared to what gets eventually verified.
	if err := txb.SetSignatures(
		signing.SignatureV2{
			PubKey: privVal.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode: signMode,
			},
		},
	); err != nil {
		panic(err)
	}

	// Generate bytes to be signed.
	bytesToSign, err := txConf.SignModeHandler().GetSignBytes(
		signing.SignMode_SIGN_MODE_DIRECT,
		authsigning.SignerData{
			ChainID: b.chainID,
			PubKey:  privVal.PubKey(),
			Address: sdk.MustBech32ifyAddressBytes("cosmos", privVal.PubKey().Address()), // TODO: don't hardcode cosmos1!

			// No account or sequence number for gentx.
		},
		txb.GetTx(),
	)
	if err != nil {
		panic(err)
	}

	// Produce the signature.
	signed, err := privVal.Sign(bytesToSign)
	if err != nil {
		panic(err)
	}

	// Set the signature on the builder.
	if err := txb.SetSignatures(
		signing.SignatureV2{
			PubKey: privVal.PubKey(),
			Data: &signing.SingleSignatureData{
				SignMode:  signMode,
				Signature: signed,
			},
		},
	); err != nil {
		panic(err)
	}

	b.gentxs = append(b.gentxs, txb.GetTx())

	return b
}

// ChainID sets the genesis's "chain_id" field.
func (b *GenesisBuilder) ChainID(id string) *GenesisBuilder {
	b.chainID = id

	var err error
	b.outer["chain_id"], err = json.Marshal(id)
	if err != nil {
		panic(err)
	}

	return b
}

// GenesisTime sets the genesis's "genesis_time" field.
// Note that [NewGenesisBuilder] sets the genesis time to the current time by default.
func (b *GenesisBuilder) GenesisTime(t time.Time) *GenesisBuilder {
	var err error
	b.outer["genesis_time"], err = json.Marshal(t.Format(time.RFC3339Nano))
	if err != nil {
		panic(err)
	}
	return b
}

// InitialHeight sets the genesis's "initial_height" field to h.
// Note that [NewGenesisBuilder] sets the initial height to 1 by default.
func (b *GenesisBuilder) InitialHeight(h int64) *GenesisBuilder {
	var err error
	b.outer["initial_height"], err = json.Marshal(strconv.FormatInt(h, 10))
	if err != nil {
		panic(err)
	}
	return b
}

// AuthParams sets the auth params on the genesis.
func (b *GenesisBuilder) AuthParams(params authtypes.Params) *GenesisBuilder {
	var err error
	b.appState[authtypes.ModuleName], err = json.Marshal(map[string]any{
		"params": params,
	})
	if err != nil {
		panic(err)
	}

	return b
}

// DefaultAuthParams calls b.AuthParams with [authtypes.DefaultParams],
// as a convenience so that callers do not have to import the authtypes package.
func (b *GenesisBuilder) DefaultAuthParams() *GenesisBuilder {
	return b.AuthParams(authtypes.DefaultParams())
}

// Consensus sets the consensus parameters and initial validators.
//
// If params is nil, [cmttypes.DefaultConsensusParams] is used.
func (b *GenesisBuilder) Consensus(params *cmttypes.ConsensusParams, vals CometGenesisValidators) *GenesisBuilder {
	if params == nil {
		params = cmttypes.DefaultConsensusParams()
	}

	var err error
	b.outer[consensusparamtypes.ModuleName], err = (&genutiltypes.ConsensusGenesis{
		Params:     params,
		Validators: vals.ToComet(),
	}).MarshalJSON()
	if err != nil {
		panic(err)
	}

	return b
}

// Staking sets the staking parameters, validators, and delegations on the genesis.
//
// This also modifies the bank state's balances to include the bonded pool balance.
func (b *GenesisBuilder) Staking(
	params stakingtypes.Params,
	vals StakingValidators,
	delegations []stakingtypes.Delegation,
) *GenesisBuilder {
	var err error
	b.appState[stakingtypes.ModuleName], err = b.codec.MarshalJSON(
		stakingtypes.NewGenesisState(params, vals.ToStakingType(), delegations),
	)
	if err != nil {
		panic(err)
	}

	// Modify bank state for bonded pool.

	var coins sdk.Coins
	for _, v := range vals {
		coins = coins.Add(sdk.NewCoin(sdk.DefaultBondDenom, v.V.Tokens))
	}

	bondedPoolBalance := banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   coins,
	}

	// get bank types genesis, add account

	bankGenesis := banktypes.GetGenesisStateFromAppState(b.codec, b.appState)
	bankGenesis.Balances = append(bankGenesis.Balances, bondedPoolBalance)

	b.appState[banktypes.ModuleName], err = b.codec.MarshalJSON(bankGenesis)
	if err != nil {
		panic(err)
	}

	return b
}

// StakingWithDefaultParams calls b.Staking, providing [stakingtypes.DefaultParams]
// so that callers don't necessarily have to import [stakingtypes].
func (b *GenesisBuilder) StakingWithDefaultParams(vals StakingValidators, delegations []stakingtypes.Delegation) *GenesisBuilder {
	return b.Staking(stakingtypes.DefaultParams(), vals, delegations)
}

// DefaultStaking is shorthand for b.StakingWithDefaultParams with nil validators and delegations.
func (b *GenesisBuilder) DefaultStaking() *GenesisBuilder {
	return b.StakingWithDefaultParams(nil, nil)
}

// Banking sets the banking genesis state.
func (b *GenesisBuilder) Banking(
	params banktypes.Params,
	balances []banktypes.Balance,
	totalSupply sdk.Coins,
	denomMetadata []banktypes.Metadata,
	sendEnabled []banktypes.SendEnabled,
) *GenesisBuilder {
	var err error
	b.appState[banktypes.ModuleName], err = b.codec.MarshalJSON(
		banktypes.NewGenesisState(
			params,
			balances,
			totalSupply,
			denomMetadata,
			sendEnabled,
		),
	)
	if err != nil {
		panic(err)
	}
	return b
}

// BankingWithDefaultParams calls b.Banking with [banktypes.DefaultParams],
// so that callers don't necessarily have to import [banktypes].
func (b *GenesisBuilder) BankingWithDefaultParams(
	balances []banktypes.Balance,
	totalSupply sdk.Coins,
	denomMetadata []banktypes.Metadata,
	sendEnabled []banktypes.SendEnabled,
) *GenesisBuilder {
	return b.Banking(
		banktypes.DefaultParams(),
		balances,
		totalSupply,
		denomMetadata,
		sendEnabled,
	)
}

// Mint sets the mint genesis state.
func (b *GenesisBuilder) Mint(m minttypes.Minter, p minttypes.Params) *GenesisBuilder {
	var err error
	b.appState[minttypes.ModuleName], err = b.codec.MarshalJSON(
		minttypes.NewGenesisState(m, p),
	)
	if err != nil {
		panic(err)
	}
	return b
}

// DefaultMint calls b.Mint with [minttypes.DefaultInitialMinter] and [minttypes.DefaultParams].
func (b *GenesisBuilder) DefaultMint() *GenesisBuilder {
	return b.Mint(minttypes.DefaultInitialMinter(), minttypes.DefaultParams())
}

// Slashing sets the slashing genesis state.
func (b *GenesisBuilder) Slashing(
	params slashingtypes.Params,
	si []slashingtypes.SigningInfo,
	mb []slashingtypes.ValidatorMissedBlocks,
) *GenesisBuilder {
	var err error
	b.appState[slashingtypes.ModuleName], err = b.codec.MarshalJSON(
		slashingtypes.NewGenesisState(params, si, mb),
	)
	if err != nil {
		panic(err)
	}
	return b
}

// SlashingWithDefaultParams calls b.Slashing with [slashingtypes.DefaultParams],
// so that callers don't necessarily have to import [slashingtypes].
func (b *GenesisBuilder) SlashingWithDefaultParams(
	si []slashingtypes.SigningInfo,
	mb []slashingtypes.ValidatorMissedBlocks,
) *GenesisBuilder {
	return b.Slashing(slashingtypes.DefaultParams(), si, mb)
}

// DefaultSlashing is shorthand for b.SlashingWithDefaultParams
// with nil signing info and validator missed blocks.
func (b *GenesisBuilder) DefaultSlashing() *GenesisBuilder {
	return b.SlashingWithDefaultParams(nil, nil)
}

// BaseAccounts sets the initial base accounts and balances.
func (b *GenesisBuilder) BaseAccounts(ba BaseAccounts, balances []banktypes.Balance) *GenesisBuilder {
	// Logic mostly copied from AddGenesisAccount.

	authGenState := authtypes.GetGenesisStateFromAppState(b.codec, b.appState)
	bankGenState := banktypes.GetGenesisStateFromAppState(b.codec, b.appState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		panic(err)
	}

	for _, a := range ba {
		accs = append(accs, a)
	}
	accs = authtypes.SanitizeGenesisAccounts(accs)

	genAccs, err := authtypes.PackAccounts(accs)
	if err != nil {
		panic(err)
	}

	authGenState.Accounts = genAccs
	jAuthGenState, err := b.codec.MarshalJSON(&authGenState)
	if err != nil {
		panic(err)
	}
	b.appState[authtypes.ModuleName] = jAuthGenState

	bankGenState.Balances = append(bankGenState.Balances, balances...)
	bankGenState.Balances = banktypes.SanitizeGenesisBalances(bankGenState.Balances)

	jBankState, err := b.codec.MarshalJSON(bankGenState)
	if err != nil {
		panic(err)
	}
	b.appState[banktypes.ModuleName] = jBankState
	return b
}

func (b *GenesisBuilder) Distribution(g *distributiontypes.GenesisState) *GenesisBuilder {
	j, err := b.codec.MarshalJSON(g)
	if err != nil {
		panic(err)
	}

	b.appState[distributiontypes.ModuleName] = j
	return b
}

func (b *GenesisBuilder) DefaultDistribution() *GenesisBuilder {
	return b.Distribution(distributiontypes.DefaultGenesisState())
}

// JSON returns the map of the genesis after applying some final transformations.
func (b *GenesisBuilder) JSON() map[string]json.RawMessage {
	gentxGenesisState := genutiltypes.NewGenesisStateFromTx(
		authtx.NewTxConfig(b.codec, tx.DefaultSignModes).TxJSONEncoder(),
		b.gentxs,
	)

	if err := genutiltypes.ValidateGenesis(
		gentxGenesisState,
		authtx.NewTxConfig(b.codec, tx.DefaultSignModes).TxJSONDecoder(),
		genutiltypes.DefaultMessageValidator,
	); err != nil {
		panic(err)
	}

	b.appState = genutiltypes.SetGenesisStateInAppState(
		b.codec, b.appState, gentxGenesisState,
	)

	appState, err := b.amino.MarshalJSON(b.appState)
	if err != nil {
		panic(err)
	}

	b.outer["app_state"] = appState

	return b.outer
}

// Encode returns the JSON-encoded, finalized genesis.
func (b *GenesisBuilder) Encode() []byte {
	j, err := b.amino.MarshalJSON(b.JSON())
	if err != nil {
		panic(err)
	}

	return j
}

// DefaultGenesisBuilderOnlyValidators returns a GenesisBuilder configured only with the given StakingValidators,
// with default parameters everywhere else.
// validatorAmount is the amount to give each validator during gentx.
//
// This is a convenience for the common case of nothing special in the genesis.
// For anything outside of the defaults,
// the longhand form of NewGenesisBuilder().ChainID(chainID)... should be used.
func DefaultGenesisBuilderOnlyValidators(
	chainID string,
	sv StakingValidators,
	validatorAmount sdk.Coin,
) *GenesisBuilder {
	cmtVals := make(CometGenesisValidators, len(sv))
	for i := range sv {
		cmtVals[i] = sv[i].C
	}

	b := NewGenesisBuilder().
		ChainID(chainID).
		DefaultAuthParams().
		Consensus(nil, cmtVals).
		BaseAccounts(sv.BaseAccounts(), nil).
		StakingWithDefaultParams(nil, nil).
		BankingWithDefaultParams(sv.Balances(), nil, nil, nil).
		DefaultDistribution().
		DefaultMint().
		SlashingWithDefaultParams(nil, nil)

	for _, v := range sv {
		b.GenTx(*v.PK.Del, v.C.V, sdk.NewCoin(sdk.DefaultBondDenom, sdk.DefaultPowerReduction))
	}

	return b
}
