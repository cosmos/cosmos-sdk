package common

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
)

func TestGaiaCLI(t *testing.T) {

	// clear genesis/keys
	_ = tests.ExecuteT(t, "gaiad unsafe_reset_all")
	_, wc0, _ := tests.GoExecuteT(t, "gaiacli keys delete foo")
	defer wc0.Close()
	_, err := wc0.Write([]byte("1234567890\n"))
	require.NoError(t, err)
	_, wc1, _ := tests.GoExecuteT(t, "gaiacli keys delete bar")
	defer wc1.Close()
	_, err = wc1.Write([]byte("1234567890\n"))
	require.NoError(t, err)
	time.Sleep(time.Second)

	// init genesis get master key
	out := tests.ExecuteT(t, "gaiad init")
	var initRes map[string]interface{}
	outCut := "{" + strings.SplitN(out, "{", 2)[1]
	err = json.Unmarshal([]byte(outCut), &initRes)
	require.NoError(t, err, "out %v outCut %v err %v", out, outCut, err)
	masterKey := (initRes["secret"]).(string)
	chainID := (initRes["chain_id"]).(string)

	// start gaiad server
	_, wc2, _ := tests.GoExecuteT(t, "gaiad start")
	defer wc2.Close()
	time.Sleep(time.Second)

	// add the master key
	_, wc3, _ := tests.GoExecuteT(t, "gaiacli keys add foo --recover")
	defer wc3.Close()
	_, err = wc3.Write([]byte("1234567890\n"))
	require.NoError(t, err)
	_, err = wc3.Write([]byte(masterKey + "\n"))
	require.NoError(t, err)
	time.Sleep(time.Second)

	// add a secondary key
	_, wc4, _ := tests.GoExecuteT(t, "gaiacli keys add bar")
	time.Sleep(time.Second * 5)
	_, err = wc4.Write([]byte("1234567890\n"))
	require.NoError(t, err)
	time.Sleep(time.Second * 5)

	// get addresses
	out = tests.ExecuteT(t, "gaiacli keys show foo")
	fooAddr := strings.TrimLeft(out, "foo\t")
	out = tests.ExecuteT(t, "gaiacli keys show bar")
	barAddr := strings.TrimLeft(out, "bar\t")
	fmt.Printf("debug barAddr: %v\n", barAddr)

	// send money from foo to bar
	cmdStr := fmt.Sprintf("gaiacli send --sequence=0 --chain-id=%v --amount=10fermion --to=%v --name=foo", chainID, barAddr)
	_, wc5, rc5 := tests.GoExecuteT(t, cmdStr)
	_, err = wc5.Write([]byte("1234567890\n"))
	require.NoError(t, err)
	fmt.Printf("debug outCh: %v\n", out)
	time.Sleep(time.Second)
	bz := make([]byte, 1000000)
	rc5.Read(bz)
	fmt.Printf("debug ex: %v\n", string(bz))

	// verify money sent to bar
	time.Sleep(time.Second)
	out = tests.ExecuteT(t, fmt.Sprintf("gaiacli account %v", fooAddr))
	fmt.Printf("debug out: %v\n", out)
	out = tests.ExecuteT(t, fmt.Sprintf("gaiacli account %v", barAddr))
	require.Fail(t, "debug out: %v\n", out)
}
