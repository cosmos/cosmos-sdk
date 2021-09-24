package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVestingAccountMsg(t *testing.T) {
	_, _, fromAddr := KeyTestPubAddr()
	_, _, toAddr := KeyTestPubAddr()
	amount := NewTestCoins()
	endTime := int64(123456789)
	msg := NewMsgCreateVestingAccount(fromAddr, toAddr, amount, endTime, false)
	require.NotNil(t, msg)
	route := msg.Route()
	require.Equal(t, RouterKey, route)
	tp := msg.Type()
	require.Equal(t, TypeMsgCreateVestingAccount, tp)
	err := msg.ValidateBasic()
	require.Nil(t, err)

	badFromMsg := MsgCreateVestingAccount{
		FromAddress: "foo",
		ToAddress:   toAddr.String(),
		Amount:      amount,
		EndTime:     endTime,
	}
	err = badFromMsg.ValidateBasic()
	require.Error(t, err)

	badToMsg := MsgCreateVestingAccount{
		FromAddress: fromAddr.String(),
		ToAddress:   "foo",
		Amount:      amount,
		EndTime:     endTime,
	}
	err = badToMsg.ValidateBasic()
	require.Error(t, err)

	badEndTime := NewMsgCreateVestingAccount(fromAddr, toAddr, amount, int64(-1), false)
	err = badEndTime.ValidateBasic()
	require.Error(t, err)
}

func TestPeriodicVestingAccountMsg(t *testing.T) {
	_, _, fromAddr := KeyTestPubAddr()
	_, _, toAddr := KeyTestPubAddr()
	amount := NewTestCoins()
	startTime := int64(123456789)
	periods := []Period{
		{Length: 86400, Amount: amount},
	}
	msg := NewMsgCreatePeriodicVestingAccount(fromAddr, toAddr, startTime, periods)
	route := msg.Route()
	require.Equal(t, RouterKey, route)
	tp := msg.Type()
	require.Equal(t, TypeMsgCreatePeriodicVestingAccount, tp)
	err := msg.ValidateBasic()
	require.Nil(t, err)

	badFromMsg := MsgCreatePeriodicVestingAccount{
		FromAddress:    "foo",
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		VestingPeriods: periods,
	}
	err = badFromMsg.ValidateBasic()
	require.Error(t, err)

	badToMsg := MsgCreatePeriodicVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      "foo",
		StartTime:      startTime,
		VestingPeriods: periods,
	}
	err = badToMsg.ValidateBasic()
	require.Error(t, err)

	badPeriods := NewMsgCreatePeriodicVestingAccount(fromAddr, toAddr, startTime, []Period{
		{Length: 0, Amount: amount},
	})
	err = badPeriods.ValidateBasic()
	require.Error(t, err)
}
