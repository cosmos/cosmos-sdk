package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func TestGaiaCLI(t *testing.T) {

	tests.ExecuteT(t, "gaiad unsafe_reset_all")
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
	time.Sleep(time.Second) // waiting for some blocks to pass

	//executeWrite(t, "gaiacli keys add foo --recover", pass, masterKey)
	cmd, wc3, _ := tests.GoExecuteT(t, "gaiacli keys add foo --recover")
	time.Sleep(time.Second) // waiting for some blocks to pass
	_, err := wc3.Write([]byte("1234567890\n"))
	require.NoError(t, err)
	_, err = wc3.Write([]byte(masterKey + "\n"))
	require.NoError(t, err)
	cmd.Wait()
	time.Sleep(time.Second * 5) // waiting for some blocks to pass
	fooAddr := executeGetAddr(t, "gaiacli keys show foo")
	panic(fmt.Sprintf("debug fooAddr: %v\n", fooAddr))

	executeWrite(t, "gaiacli keys add bar", pass)
	barAddr := executeGetAddr(t, "gaiacli keys show bar")
	executeWrite(t, fmt.Sprintf("gaiacli send %v --amount=10fermion --to=%v --name=foo", flags, barAddr), pass)
	time.Sleep(time.Second * 3) // waiting for some blocks to pass

	barAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", barAddr, flags))
	assert.Equal(t, int64(10), barAcc.GetCoins().AmountOf("fermion"))
	fooAcc := executeGetAccount(t, fmt.Sprintf("gaiacli account %v %v", fooAddr, flags))
	assert.Equal(t, int64(99990), fooAcc.GetCoins().AmountOf("fermion"))

	// declare candidacy
	//executeWrite(t, "gaiacli declare-candidacy -", pass)
}

func executeWrite(t *testing.T, cmdStr string, writes ...string) {
	cmd, wc, _ := tests.GoExecuteT(t, cmdStr)
	for _, write := range writes {
		_, err := wc.Write([]byte(write + "\n"))
		require.NoError(t, err)
	}
	cmd.Wait()
}

func executeInit(t *testing.T, cmdStr string) (masterKey, chainID string) {
	tests.GoExecuteT(t, cmdStr)
	out := tests.ExecuteT(t, cmdStr)
	outCut := "{" + strings.SplitN(out, "{", 2)[1] // weird I'm sorry

	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(outCut), &initRes)
	require.NoError(t, err, "out %v outCut %v err %v", out, outCut, err)
	masterKey = string(initRes["secret"])
	chainID = string(initRes["chain_id"])
	return
}

func executeGetAddr(t *testing.T, cmdStr string) (addr string) {
	out := tests.ExecuteT(t, cmdStr)
	name := strings.SplitN(cmdStr, " show ", 2)[1]
	return strings.TrimLeft(out, name+"\t")
}

func executeGetAccount(t *testing.T, cmdStr string) auth.BaseAccount {
	out := tests.ExecuteT(t, cmdStr)
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(out), &initRes)
	require.NoError(t, err, "out %v, err %v", out, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	_ = json.Unmarshal(value, &acc) //XXX pubkey can't be decoded go amino issue
	require.NoError(t, err, "value %v, err %v", string(value), err)
	return acc
}
