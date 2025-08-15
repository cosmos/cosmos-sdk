package simapp

import (
	"context"
	"errors"
	"testing"
	"time"

	storeaddress "cosmossdk.io/core/address"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txsigning "cosmossdk.io/x/tx/signing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// dummy implementations for required interfaces used only for constructing the ante handler chain.

type dummyAddressCodec struct{}

func (d dummyAddressCodec) StringToBytes(text string) ([]byte, error) { return nil, errors.New("not implemented") }
func (d dummyAddressCodec) BytesToString(bz []byte) (string, error)   { return "", errors.New("not implemented") }

type dummyAccountKeeper struct{}

func (d dummyAccountKeeper) GetParams(ctx context.Context) (params authtypes.Params) { return authtypes.Params{} }
func (d dummyAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}
func (d dummyAccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI)                 {}
func (d dummyAccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress               { return nil }
func (d dummyAccountKeeper) AddressCodec() storeaddress.Codec                               { return dummyAddressCodec{} }
func (d dummyAccountKeeper) UnorderedTransactionsEnabled() bool                             { return false }
func (d dummyAccountKeeper) RemoveExpiredUnorderedNonces(ctx sdk.Context) error             { return nil }
func (d dummyAccountKeeper) TryAddUnorderedNonce(ctx sdk.Context, sender []byte, t time.Time) error {
	return nil
}

type dummyBankKeeper struct{}

func (d dummyBankKeeper) IsSendEnabledCoins(ctx context.Context, coins ...sdk.Coin) error { return nil }
func (d dummyBankKeeper) SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error {
	return nil
}
func (d dummyBankKeeper) SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	return nil
}

type dummyFeegrantKeeper struct{}

func (d dummyFeegrantKeeper) UseGrantedFees(ctx context.Context, granter, grantee sdk.AccAddress, fee sdk.Coins, msgs []sdk.Msg) error {
	return nil
}

type dummyCircuitBreaker struct{}

func (d dummyCircuitBreaker) IsAllowed(ctx context.Context, typeURL string) (bool, error) { return true, nil }

type dummySignModeHandler struct{}

func (d dummySignModeHandler) Mode() signingv1beta1.SignMode { return signingv1beta1.SignMode_SIGN_MODE_DIRECT }
func (d dummySignModeHandler) GetSignBytes(_ context.Context, _ txsigning.SignerData, _ txsigning.TxData) ([]byte, error) {
	return nil, nil
}

func makeBaseHandlerOptions() ante.HandlerOptions {
	return ante.HandlerOptions{
		AccountKeeper:   dummyAccountKeeper{},
		BankKeeper:      dummyBankKeeper{},
		FeegrantKeeper:  dummyFeegrantKeeper{},
		SignModeHandler: txsigning.NewHandlerMap(dummySignModeHandler{}),
		SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
	}
}

func TestNewAnteHandler_CircuitKeeperNil(t *testing.T) {
	_, err := NewAnteHandler(HandlerOptions{
		HandlerOptions: makeBaseHandlerOptions(),
		CircuitKeeper:  nil,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "circuit keeper is required for ante builder")
}

func TestNewAnteHandler_CircuitKeeperProvided(t *testing.T) {
	h, err := NewAnteHandler(HandlerOptions{
		HandlerOptions: makeBaseHandlerOptions(),
		CircuitKeeper:  dummyCircuitBreaker{},
	})
	require.NoError(t, err)
	require.NotNil(t, h)
}
