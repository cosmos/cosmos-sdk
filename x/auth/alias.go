package auth

import (
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// DONTCOVER
// nolint

const (
	ModuleName                    = types.ModuleName
	StoreKey                      = types.StoreKey
	FeeCollectorName              = types.FeeCollectorName
	QuerierRoute                  = types.QuerierRoute
	DefaultParamspace             = types.DefaultParamspace
	DefaultMaxMemoCharacters      = types.DefaultMaxMemoCharacters
	DefaultTxSigLimit             = types.DefaultTxSigLimit
	DefaultTxSizeCostPerByte      = types.DefaultTxSizeCostPerByte
	DefaultSigVerifyCostED25519   = types.DefaultSigVerifyCostED25519
	DefaultSigVerifyCostSecp256k1 = types.DefaultSigVerifyCostSecp256k1
	QueryAccount                  = types.QueryAccount
	QueryParams                   = types.QueryParams
	MaxGasWanted                  = types.MaxGasWanted
	Minter                        = types.Minter
	Burner                        = types.Burner
	Staking                       = types.Staking
)

var (
	// functions aliases
	NewAnteHandler                    = ante.NewAnteHandler
	GetSignerAcc                      = ante.GetSignerAcc
	DefaultSigVerificationGasConsumer = ante.DefaultSigVerificationGasConsumer
	DeductFees                        = ante.DeductFees
	SetGasMeter                       = ante.SetGasMeter
	NewAccountKeeper                  = keeper.NewAccountKeeper
	NewQuerier                        = keeper.NewQuerier
	NewBaseAccount                    = types.NewBaseAccount
	ProtoBaseAccount                  = types.ProtoBaseAccount
	NewBaseAccountWithAddress         = types.NewBaseAccountWithAddress
	NewAccountRetriever               = types.NewAccountRetriever
	RegisterCodec                     = types.RegisterCodec
	NewGenesisState                   = types.NewGenesisState
	DefaultGenesisState               = types.DefaultGenesisState
	ValidateGenesis                   = types.ValidateGenesis
	SanitizeGenesisAccounts           = types.SanitizeGenesisAccounts
	AddressStoreKey                   = types.AddressStoreKey
	NewParams                         = types.NewParams
	ParamKeyTable                     = types.ParamKeyTable
	DefaultParams                     = types.DefaultParams
	NewQueryAccountParams             = types.NewQueryAccountParams
	NewStdTx                          = types.NewStdTx
	CountSubKeys                      = types.CountSubKeys
	NewStdFee                         = types.NewStdFee //nolint:staticcheck
	StdSignBytes                      = types.StdSignBytes
	DefaultTxDecoder                  = types.DefaultTxDecoder
	DefaultTxEncoder                  = types.DefaultTxEncoder
	NewTxBuilder                      = types.NewTxBuilder
	NewTxBuilderFromCLI               = types.NewTxBuilderFromCLI
	MakeSignature                     = types.MakeSignature
	ValidateGenAccounts               = types.ValidateGenAccounts
	GetGenesisStateFromAppState       = types.GetGenesisStateFromAppState
	NewStdSignature                   = types.NewStdSignature
	NewModuleAddress                  = types.NewModuleAddress
	NewEmptyModuleAccount             = types.NewEmptyModuleAccount
	NewModuleAccount                  = types.NewModuleAccount

	// variable aliases
	AddressStoreKeyPrefix     = types.AddressStoreKeyPrefix
	GlobalAccountNumberKey    = types.GlobalAccountNumberKey
	KeyMaxMemoCharacters      = types.KeyMaxMemoCharacters
	KeyTxSigLimit             = types.KeyTxSigLimit
	KeyTxSizeCostPerByte      = types.KeyTxSizeCostPerByte
	KeySigVerifyCostED25519   = types.KeySigVerifyCostED25519
	KeySigVerifyCostSecp256k1 = types.KeySigVerifyCostSecp256k1
)

type (
	AccountI                         = types.AccountI
	SignatureVerificationGasConsumer = ante.SignatureVerificationGasConsumer
	AccountKeeper                    = keeper.AccountKeeper
	BaseAccount                      = types.BaseAccount
	AccountRetriever                 = types.AccountRetriever
	GenesisState                     = types.GenesisState
	Params                           = types.Params
	QueryAccountParams               = types.QueryAccountParams
	StdSignMsg                       = types.StdSignMsg
	StdTx                            = types.StdTx
	StdFee                           = types.StdFee //nolint:staticcheck
	StdSignDoc                       = types.StdSignDoc
	StdSignature                     = types.StdSignature //nolint:staticcheck
	TxBuilder                        = types.TxBuilder
	GenesisAccountIterator           = types.GenesisAccountIterator
	ModuleAccount                    = types.ModuleAccount
	GenesisAccounts                  = types.GenesisAccounts
	GenesisAccount                   = types.GenesisAccount
)
