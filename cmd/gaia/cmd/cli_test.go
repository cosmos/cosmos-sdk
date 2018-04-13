package common

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

func password(t *testing.T, wc io.WriteCloser) {
	_, err := wc.Write([]byte("1234567890\n"))
	require.NoError(t, err)
}

func TestGaiaCLI(t *testing.T) {

	// clear genesis/keys
	tests.ExecuteT(t, "gaiad unsafe_reset_all")
	cmd, wc0, _ := tests.GoExecuteT(t, "gaiacli keys delete foo")
	password(t, wc0)
	cmd.Wait()
	cmd, wc1, _ := tests.GoExecuteT(t, "gaiacli keys delete bar")
	password(t, wc1)
	cmd.Wait()

	// init genesis get master key
	out := tests.ExecuteT(t, "gaiad init")
	var initRes map[string]interface{}
	outCut := "{" + strings.SplitN(out, "{", 2)[1] // weird I'm sorry
	err := json.Unmarshal([]byte(outCut), &initRes)
	require.NoError(t, err, "out %v outCut %v err %v", out, outCut, err)
	masterKey := (initRes["secret"]).(string)
	chainID := (initRes["chain_id"]).(string)

	servAddr := server.FreeTCPAddr(t)
	gaiacliFlags := fmt.Sprintf("--node=%v --chain-id=%v", servAddr, chainID)

	// start gaiad server
	cmd, _, _ = tests.GoExecuteT(t, fmt.Sprintf("gaiad start --rpc.laddr=%v", servAddr))
	defer cmd.Process.Kill()

	// add the master key
	cmd, wc3, _ := tests.GoExecuteT(t, "gaiacli keys add foo --recover")
	password(t, wc3)
	_, err = wc3.Write([]byte(masterKey + "\n"))
	require.NoError(t, err)
	cmd.Wait()

	// add a secondary key
	cmd, wc4, _ := tests.GoExecuteT(t, "gaiacli keys add bar")
	password(t, wc4)
	cmd.Wait()

	// get addresses
	out = tests.ExecuteT(t, "gaiacli keys show foo")
	fooAddr := strings.TrimLeft(out, "foo\t")
	out = tests.ExecuteT(t, "gaiacli keys show bar")
	barAddr := strings.TrimLeft(out, "bar\t")

	// send money from foo to bar
	cmdStr := fmt.Sprintf("gaiacli send %v --sequence=0 --amount=10fermion --to=%v --name=foo", gaiacliFlags, barAddr)
	cmd, wc5, _ := tests.GoExecuteT(t, cmdStr)
	password(t, wc5)
	cmd.Wait()
	time.Sleep(time.Second * 3) // waiting for some blocks to pass

	// verify money sent to bar
	out = tests.ExecuteT(t, fmt.Sprintf("gaiacli account %v %v", barAddr, gaiacliFlags))
	barAcc := unmarshalBaseAccount(t, out)
	assert.Equal(t, int64(10), barAcc.GetCoins().AmountOf("fermion"))

	out = tests.ExecuteT(t, fmt.Sprintf("gaiacli account %v %v", fooAddr, gaiacliFlags))
	fooAcc := unmarshalBaseAccount(t, out)
	assert.Equal(t, int64(99990), fooAcc.GetCoins().AmountOf("fermion"))
}

func unmarshalBaseAccount(t *testing.T, raw string) auth.BaseAccount {
	var initRes map[string]json.RawMessage
	err := json.Unmarshal([]byte(raw), &initRes)
	require.NoError(t, err, "raw %v, err %v", raw, err)
	value := initRes["value"]
	var acc auth.BaseAccount
	_ = json.Unmarshal(value, &acc) //XXX pubkey can't be decoded go amino issue
	require.NoError(t, err, "value %v, err %v", string(value), err)
	return acc
}
