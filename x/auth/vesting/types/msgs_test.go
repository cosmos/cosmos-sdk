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
	require.NoError(t, err)

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
	msg := NewMsgCreatePeriodicVestingAccount(fromAddr, toAddr, startTime, periods, false)
	route := msg.Route()
	require.Equal(t, RouterKey, route)
	tp := msg.Type()
	require.Equal(t, TypeMsgCreatePeriodicVestingAccount, tp)
	err := msg.ValidateBasic()
	require.NoError(t, err)

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
	}, false)
	err = badPeriods.ValidateBasic()
	require.Error(t, err)
}

func TestClawbackVestingAccountMsg(t *testing.T) {
	_, _, fromAddr := KeyTestPubAddr()
	_, _, toAddr := KeyTestPubAddr()
	amount := NewTestCoins()
	startTime := int64(100200300)
	lockupPeriods := []Period{
		{Length: 200000, Amount: amount},
	}
	vestingPeriods := []Period{
		{Length: 300000, Amount: amount},
	}
	msg := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, lockupPeriods, vestingPeriods, false)
	route := msg.Route()
	require.Equal(t, RouterKey, route)
	tp := msg.Type()
	require.Equal(t, TypeMsgCreateClawbackVestingAccount, tp)
	err := msg.ValidateBasic()
	require.NoError(t, err)

	badFromMsg := MsgCreateClawbackVestingAccount{
		FromAddress:    "foo",
		ToAddress:      toAddr.String(),
		StartTime:      startTime,
		LockupPeriods:  lockupPeriods,
		VestingPeriods: vestingPeriods,
	}
	err = badFromMsg.ValidateBasic()
	require.Error(t, err)

	badToMsg := MsgCreateClawbackVestingAccount{
		FromAddress:    fromAddr.String(),
		ToAddress:      "foo",
		StartTime:      startTime,
		LockupPeriods:  lockupPeriods,
		VestingPeriods: vestingPeriods,
	}
	err = badToMsg.ValidateBasic()
	require.Error(t, err)

	badPeriods := []Period{{Length: 0, Amount: amount}}
	badLockup := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, badPeriods, vestingPeriods, false)
	err = badLockup.ValidateBasic()
	require.Error(t, err)

	badVesting := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, lockupPeriods, badPeriods, false)
	err = badVesting.ValidateBasic()
	require.Error(t, err)

	badAmounts := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, lockupPeriods, []Period{
		{Length: 17, Amount: amount.Add(amount...)},
	}, false)
	err = badAmounts.ValidateBasic()
	require.Error(t, err)

	emptyPeriods := []Period{}
	noLockupOk := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, emptyPeriods, vestingPeriods, false)
	err = noLockupOk.ValidateBasic()
	require.NoError(t, err)

	noVestingOk := NewMsgCreateClawbackVestingAccount(fromAddr, toAddr, startTime, lockupPeriods, emptyPeriods, false)
	err = noVestingOk.ValidateBasic()
	require.NoError(t, err)
}

func TestClawbackMsg(t *testing.T) {
	_, _, funderAddr := KeyTestPubAddr()
	_, _, addr := KeyTestPubAddr()
	_, _, destAddr := KeyTestPubAddr()

	okMsg := NewMsgClawback(funderAddr, addr, destAddr)
	route := okMsg.Route()
	require.Equal(t, RouterKey, route)
	tp := okMsg.Type()
	require.Equal(t, TypeMsgClawback, tp)
	err := okMsg.ValidateBasic()
	require.NoError(t, err)

	noDest := NewMsgClawback(funderAddr, addr, nil)
	require.Equal(t, noDest.DestAddress, "")
	err = noDest.ValidateBasic()
	require.NoError(t, err)

	badFunder := MsgClawback{
		FunderAddress: "foo",
		Address:       addr.String(),
		DestAddress:   destAddr.String(),
	}
	err = badFunder.ValidateBasic()
	require.Error(t, err)

	badAddr := MsgClawback{
		FunderAddress: funderAddr.String(),
		Address:       "foo",
		DestAddress:   destAddr.String(),
	}
	err = badAddr.ValidateBasic()
	require.Error(t, err)

	badDest := MsgClawback{
		FunderAddress: funderAddr.String(),
		Address:       addr.String(),
		DestAddress:   "foo",
	}
	err = badDest.ValidateBasic()
	require.Error(t, err)
}
