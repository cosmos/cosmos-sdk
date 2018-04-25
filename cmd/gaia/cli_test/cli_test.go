package clitest

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

func TestGaiaCLISend(t *testing.T) {

	tests.ExecuteT(t, "gaiad unsafe_reset_all", 1)
	pass := "1234567890"
	executeWrite(t, "gaiacli keys delete foo", pass)
	executeWrite(t, "gaiacli keys delete bar", pass)
	masterKey, chainID := executeInit(t, "gaiad init")

	// get a free port, also setup some common flags
	servAddr := server.FreeTCPAddr(t)
	flags := fmt.Sprintf("--node=%v --chain-id=%v", servAddr, chainID)

	// start gaiad server
	cmd, _, _ := tests.GoExecuteT(t, fmt.Sprintf("gaiad start --rpc.laddr=%v", servAddr))
	defer cmd.Process.Kill()

	executeWrite(t, "gaiacli keys add foo --recover", pass, masterKey)
	executeWrite(t, "gaiacli keys add bar", pass)

	fooAddr, _ := executeGetAddrPK(t, "gaiacli keys show foo --output=json")
	barAddr, _ := executeGetAddrPK(t, "gaiacli keys show bar --output=json")

	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooAddr, flags))
	assert.Equal(t, int64(100000), fooAcc.GetCoins().AmountOf("fermion"))

	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10fermion --to=%v --name=foo", flags, barAddr), pass)
	time.Sleep(time.Second * 3) // waiting for some blocks to pass

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barAddr, flags))
	assert.Equal(t, int64(10), barAcc.GetCoins().AmountOf("fermion"))
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooAddr, flags))
	assert.Equal(t, int64(99990), fooAcc.GetCoins().AmountOf("fermion"))
}

func TestGaiaCLIDeclareCandidacy(t *testing.T) {

	tests.ExecuteT(t, "gaiad unsafe_reset_all", 1)
	pass := "1234567890"
	executeWrite(t, "gaiacli keys delete foo", pass)
	masterKey, chainID := executeInit(t, "gaiad init")

	// get a free port, also setup some common flags
	servAddr := server.FreeTCPAddr(t)
	flags := fmt.Sprintf("--node=%v --chain-id=%v", servAddr, chainID)

	// start gaiad server
	cmd, _, _ := tests.GoExecuteT(t, fmt.Sprintf("gaiad start --rpc.laddr=%v", servAddr))
	defer cmd.Process.Kill()

	executeWrite(t, "gaiacli keys add foo --recover", pass, masterKey)
	fooAddr, fooPubKey := executeGetAddrPK(t, "gaiacli keys show foo --output=json")
	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooAddr, flags))
	assert.Equal(t, int64(100000), fooAcc.GetCoins().AmountOf("fermion"))

	// declare candidacy
	declStr := fmt.Sprintf("gaiacli declare-candidacy %v", flags)
	declStr += fmt.Sprintf(" --name=%v", "foo")
	declStr += fmt.Sprintf(" --address-candidate=%v", fooAddr)
	declStr += fmt.Sprintf(" --pubkey=%v", fooPubKey)
	declStr += fmt.Sprintf(" --amount=%v", "3fermion")
	declStr += fmt.Sprintf(" --moniker=%v", "foo-vally")
	fmt.Printf("debug declStr: %v\n", declStr)
	executeWrite(t, declStr, pass)
	time.Sleep(time.Second * 3) // waiting for some blocks to pass
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooAddr, flags))
	assert.Equal(t, int64(99997), fooAcc.GetCoins().AmountOf("fermion"))
	candidate := executeGetCandidate(t, fmt.Sprintf("gaiacli candidate %v --address-candidate=%v", flags, fooAddr))
	assert.Equal(t, candidate.Address.String(), fooAddr)
	assert.Equal(t, int64(3), candidate.Assets.Evaluate())

	// TODO timeout issues if not connected to the internet
	// unbond a single share
	unbondStr := fmt.Sprintf("gaiacli unbond %v", flags)
	unbondStr += fmt.Sprintf(" --name=%v", "foo")
	unbondStr += fmt.Sprintf(" --address-candidate=%v", fooAddr)
	unbondStr += fmt.Sprintf(" --address-delegator=%v", fooAddr)
	unbondStr += fmt.Sprintf(" --shares=%v", "1")
	unbondStr += fmt.Sprintf(" --sequence=%v", "1")
	fmt.Printf("debug unbondStr: %v\n", unbondStr)
	executeWrite(t, unbondStr, pass)
	time.Sleep(time.Second * 3) // waiting for some blocks to pass
	fooAcc = executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooAddr, flags))
	assert.Equal(t, int64(99998), fooAcc.GetCoins().AmountOf("fermion"))
	candidate = executeGetCandidate(t, fmt.Sprintf("gaiacli candidate %v --address-candidate=%v", flags, fooAddr))
	assert.Equal(t, int64(2), candidate.Assets.Evaluate())
}

func executeWrite(t *testing.T, cmdStr string, writes ...string) {
	cmd, wc, _ := tests.GoExecuteT(t, cmdStr)

	for _, write := range writes {
		_, err := wc.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}
	cmd.Wait()
}

func executeWritePrint(t *testing.T, cmdStr string, writes ...string) {
	cmd, wc, rc := tests.GoExecuteT(t, cmdStr)

	for _, write := range writes {
		_, err := wc.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}
	cmd.Wait()

	bz := make([]byte, 100000)
	rc.Read(bz)
	fmt.Printf("debug read: %v\n", string(bz))
}

func executeInit(t *testing.T, cmdStr string) (masterKey, chainID string) {
	out := tests.ExecuteT(t, cmdStr, 1)
	outCut := "{" + strings.SplitN(out, "{", 2)[1] // weird I'm sorry

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(outCut), &initRes)
	require.NoError(t, err)
	err = json.Unmarshal(initRes["secret"], &masterKey)
	require.NoError(t, err)
	err = json.Unmarshal(initRes["chain_id"], &chainID)
	require.NoError(t, err)
	return
}

func executeGetAddrPK(t *testing.T, cmdStr string) (addr, pubKey string) {
	out := tests.ExecuteT(t, cmdStr, 2)
	var ko keys.KeyOutput
	keys.UnmarshalJSON([]byte(out), &ko)
	return ko.Address, ko.PubKey
}

func executeGetAccount(t *testing.T, cmdStr string) auth.BaseAccount {
	out := tests.ExecuteT(t, cmdStr, 2)
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	_ = json.Unmarshal(value, &acc) //XXX pubkey can't be decoded go amino issue
	require.NoError(t, err, "value %v, err %v", string(value), err)
	return acc
}

func executeGetCandidate(t *testing.T, cmdStr string) stake.Candidate {
	out := tests.ExecuteT(t, cmdStr, 2)
	var candidate stake.Candidate
	cdc := app.MakeCodec()
	err := cdc.UnmarshalJSON([]byte(out), &candidate)
	require.NoError(t, err, "out %v, err %v", out, err)
	return candidate
}
