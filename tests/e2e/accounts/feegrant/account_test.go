package feegrant

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	basev1 "cosmossdk.io/x/accounts/defaults/base/v1"
	multisigv1 "cosmossdk.io/x/accounts/defaults/multisig/v1"
	feegrantv1 "cosmossdk.io/x/accounts/extensions/feegrant/v1"
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

func TestCreateAndUseGrantWithBaseAccount(t *testing.T) {
	s := setupApp(t)

	ctx := sdk.NewContext(s.CommitMultiStore(), false, s.Logger()).WithHeaderInfo(header.Info{
		Time: time.Now(),
	})
	addressCodec := s.AuthKeeper.AddressCodec()

	granteeAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	granteeAddrStr := must(addressCodec.BytesToString(granteeAddr))

	anyAddr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	anyAddrStr := must(addressCodec.BytesToString(anyAddr))

	specs := map[string]struct {
		src func(t *testing.T, ctx context.Context) sdk.AccAddress
	}{
		"legacy base migrated": {
			src: func(t *testing.T, ctx context.Context) sdk.AccAddress {
				myGranterPub := secp256k1.GenPrivKey().PubKey()
				myGranterAddr := sdk.AccAddress(myGranterPub.Address())
				myGranterAddrStr := must(addressCodec.BytesToString(myGranterAddr))

				base := &authtypes.BaseAccount{Address: myGranterAddrStr}
				updatedMod := s.AuthKeeper.NewAccount(ctx, base)
				s.AuthKeeper.SetAccount(ctx, updatedMod)

				granterAccount := s.AuthKeeper.GetAccount(ctx, myGranterAddr)
				require.NotNil(t, t, granterAccount)
				pk, err := codectypes.NewAnyWithValue(myGranterPub)
				require.NoError(t, err)

				_, err = s.AuthKeeper.AccountsModKeeper.MigrateLegacyAccount(ctx, myGranterAddr, granterAccount.GetSequence(), "base",
					&basev1.MsgInit{PubKey: pk, InitSequence: granterAccount.GetSequence()},
				)
				require.NoError(t, err)
				return myGranterAddr
			},
		},
		"xaccounts multisig": {
			src: func(t *testing.T, ctx context.Context) sdk.AccAddress {
				sender := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
				senderStr := must(addressCodec.BytesToString(sender))
				_, accAddr, err := s.AccountsKeeper.Init(ctx, "multisig", sender,
					&multisigv1.MsgInit{
						Members: []*multisigv1.Member{{Address: senderStr, Weight: 100}},
						Config: &multisigv1.Config{
							Threshold:      100,
							Quorum:         100,
							VotingPeriod:   120,
							Revote:         false,
							EarlyExecution: true,
						},
					}, nil)
				require.NoError(t, err)
				return accAddr
			},
		},
	}
	for name, spec := range specs {
		t.Run(name, func(t *testing.T) {
			granterAddr := spec.src(t, ctx)
			_, err := s.AccountsKeeper.AccountsByType.Get(ctx, granterAddr)
			require.NoError(t, err)

			// and add allowance
			expireTime := time.Now().Add(time.Hour).UTC()
			allowance := &xfeegrant.BasicAllowance{
				SpendLimit: sdk.NewCoins(sdk.NewInt64Coin("stake", 1000)),
				Expiration: &expireTime,
			}
			// when
			err = s.FeeGrantKeeper.GrantAllowance(ctx, granterAddr, granteeAddr, allowance)
			require.NoError(t, err)
			// then
			gotAllowance, err := s.FeeGrantKeeper.GetAllowance(ctx, granterAddr, granteeAddr)
			require.NoError(t, err)
			assert.NotNil(t, gotAllowance)
			// and when allowance used
			// by allowed grantee
			anyAllowedMsgType := bank.NewMsgSend(granteeAddrStr, anyAddrStr, sdk.NewCoins(sdk.NewInt64Coin("stake", 1)))
			payload := must(feegrantv1.NewMsgUseGrantedFees(granteeAddrStr, anyAllowedMsgType))
			_, err = s.AccountsKeeper.Execute(ctx, granterAddr, granteeAddr, payload, nil)
			require.NoError(t, err)
			// by unknown grantee
			payload = must(feegrantv1.NewMsgUseGrantedFees(anyAddrStr, anyAllowedMsgType))
			_, err = s.AccountsKeeper.Execute(ctx, granterAddr, anyAddr, payload, nil)
			require.Error(t, err)
		})
	}
}
