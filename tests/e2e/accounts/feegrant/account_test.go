package feegrant

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	"cosmossdk.io/core/transaction"
	basev1 "cosmossdk.io/x/accounts/defaults/base/v1"
	v1 "cosmossdk.io/x/accounts/defaults/feegrant/v1"
	bank "cosmossdk.io/x/bank/types"
	xfeegrant "cosmossdk.io/x/feegrant"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, NewE2ETestSuite())
}

// TestSimpleSendProposal creates a multisig account with 1 member, sends a tx, votes and executes it.
func (s *E2ETestSuite) TestCreateAndUseGrantWithNewAddress() {
	ctx := sdk.NewContext(s.app.CommitMultiStore(), false, s.app.Logger()).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})

	granterAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	anyAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	addressCodec := s.app.AuthKeeper.AddressCodec()
	granteeAddrStr := must(addressCodec.BytesToString(secp256k1.GenPrivKey().PubKey().Address()))
	anyAddrStr := must(addressCodec.BytesToString(anyAddr))

	_, feeGrantAccountAddr, err := s.app.AccountsKeeper.Init(ctx, "feegrant", granterAddr, &v1.MsgInit{}, nil)
	s.NoError(err)

	expireTime := time.Now().Add(time.Hour).UTC()
	allowance := &xfeegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		Expiration: &expireTime,
	}
	var payload transaction.Msg
	payload, err = v1.NewMsgGrantAllowance(allowance, granteeAddrStr)
	s.NoError(err)
	err = s.ExecuteTX(ctx, payload, feeGrantAccountAddr, granterAddr)
	s.NoError(err)

	// when
	// allowed grantee
	anyAllowedMsgType := bank.NewMsgSend(granteeAddrStr, anyAddrStr, sdk.NewCoins(sdk.NewInt64Coin("stake", 1)))
	payload = must(v1.NewMsgUseGrantedFees(granteeAddrStr, anyAllowedMsgType))
	err = s.ExecuteTX(ctx, payload, feeGrantAccountAddr, granterAddr)
	s.NoError(err)
	// unknown grantee
	payload = must(v1.NewMsgUseGrantedFees(anyAddrStr, anyAllowedMsgType))
	err = s.ExecuteTX(ctx, payload, feeGrantAccountAddr, granterAddr)
	s.Error(err)
}

func (s *E2ETestSuite) TestCreateAndUseGrantWithBaseAccount() {
	ctx := sdk.NewContext(s.app.CommitMultiStore(), false, s.app.Logger()).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})
	addressCodec := s.app.AuthKeeper.AddressCodec()

	granterPub := secp256k1.GenPrivKey().PubKey()
	granterAddr := sdk.AccAddress(granterPub.Address())
	granteeAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	granteeAddrStr := must(addressCodec.BytesToString(granteeAddr))
	anyAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	anyAddrStr := must(addressCodec.BytesToString(anyAddr))

	granterAddrStr := must(addressCodec.BytesToString(granterAddr))

	baseAcc := &authtypes.BaseAccount{Address: granterAddrStr}
	updatedMod := s.app.AuthKeeper.NewAccount(ctx, baseAcc)
	s.app.AuthKeeper.SetAccount(ctx, updatedMod)

	granterAccount := s.app.AuthKeeper.GetAccount(ctx, granterAddr)
	s.Require().NotNil(granterAccount)
	pk, err := codectypes.NewAnyWithValue(granterPub)
	s.Require().NoError(err)

	_, err = s.app.AuthKeeper.AccountsModKeeper.MigrateLegacyAccount(ctx, granterAddr, granterAccount.GetSequence(), "base", &basev1.MsgInit{
		pk, 1,
	})
	s.Require().NoError(err)
	_, err = s.app.AccountsKeeper.AccountsByType.Get(ctx, granterAddr)
	s.Require().NoError(err)

	expireTime := time.Now().Add(time.Hour).UTC()
	allowance := &xfeegrant.BasicAllowance{
		SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
		Expiration: &expireTime,
	}
	err = s.app.FeeGrantKeeper.GrantAllowance(ctx, granterAddr, granteeAddr, allowance)
	s.Require().NoError(err)
	// when
	// allowed grantee
	anyAllowedMsgType := bank.NewMsgSend(granteeAddrStr, anyAddrStr, sdk.NewCoins(sdk.NewInt64Coin("stake", 1)))
	payload := must(v1.NewMsgUseGrantedFees(granteeAddrStr, anyAllowedMsgType))
	err = s.ExecuteTX(ctx, payload, granterAddr, granteeAddr)
	s.NoError(err)
	// unknown grantee
	payload = must(v1.NewMsgUseGrantedFees(anyAddrStr, anyAllowedMsgType))
	err = s.ExecuteTX(ctx, payload, granterAddr, anyAddr)
	s.Error(err)
}
