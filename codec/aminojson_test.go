package codec_test

import (
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/go-amino"
	"google.golang.org/protobuf/proto"

	authapi "cosmossdk.io/api/cosmos/auth/v1beta1"
	"cosmossdk.io/api/cosmos/crypto/ed25519"
	distapi "cosmossdk.io/api/cosmos/distribution/v1beta1"
	"cosmossdk.io/x/tx/aminojson"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

func TestAminoJSON_LegacyParity(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterConcrete(authtypes.Params{}, "cosmos-sdk/x/auth/Params", nil)
	cdc.RegisterConcrete(disttypes.MsgWithdrawDelegatorReward{}, "cosmos-sdk/MsgWithdrawDelegationReward", nil)
	cdc.RegisterConcrete(&ed25519.PubKey{}, cryptotypes.PubKeyName, nil)
	aj := aminojson.NewAminoJSON()

	cases := map[string]struct {
		gogo   gogoproto.Message
		pulsar proto.Message
	}{
		"auth/params": {gogo: &authtypes.Params{TxSigLimit: 10}, pulsar: &authapi.Params{TxSigLimit: 10}},
		"distribution/delegator_starting_info": {
			gogo:   &disttypes.DelegatorStartingInfo{},
			pulsar: &distapi.DelegatorStartingInfo{},
		},
		"distribution/delegator_starting_info/non_zero_dec": {
			gogo:   &disttypes.DelegatorStartingInfo{Stake: types.NewDec(10)},
			pulsar: &distapi.DelegatorStartingInfo{Stake: "10.000000000000000000"},
		},
		"distribution/delegation_delegator_reward": {
			gogo:   &disttypes.DelegationDelegatorReward{},
			pulsar: &distapi.DelegationDelegatorReward{},
		},
		"distribution/community_pool_spend_proposal_with_deposit": {
			gogo:   &disttypes.CommunityPoolSpendProposalWithDeposit{},
			pulsar: &distapi.CommunityPoolSpendProposalWithDeposit{},
		},
		"distribution/msg_withdraw_delegator_reward": {
			gogo:   &disttypes.MsgWithdrawDelegatorReward{DelegatorAddress: "foo"},
			pulsar: &distapi.MsgWithdrawDelegatorReward{DelegatorAddress: "foo"},
		},
		"crypto/pubkey": {
			gogo: &cryptotypes.PubKey{Key: []byte("key")}, pulsar: &ed25519.PubKey{Key: []byte("key")},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			gogoBytes, err := cdc.MarshalJSON(tc.gogo)
			require.NoError(t, err)

			pulsarBytes, err := aj.MarshalAmino(tc.pulsar)
			require.NoError(t, err)

			require.Equal(t, string(gogoBytes), string(pulsarBytes), "gogo: %s vs pulsar: %s", gogoBytes, pulsarBytes)
		})
	}
}
