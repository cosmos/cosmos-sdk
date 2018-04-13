package common

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
)

func TestGaiaCLI(t *testing.T) {

	_, out := tests.ExecuteT(t, "gaiad init")
	var initRes map[string]interface{}
	outCut := "{" + strings.SplitN(out, "{", 2)[1]
	err := json.Unmarshal([]byte(outCut), &initRes)
	require.NoError(t, err, "out %v outCut %v err %v", out, outCut, err)
	masterKey := (initRes["secret"]).(string)
	_ = masterKey

	//wc1, _ := tests.GoExecuteT(t, "gaiacli keys add foo --recover")
	//time.Sleep(time.Second)
	//_, err = wc1.Write([]byte("1234567890\n"))
	//time.Sleep(time.Second)
	//_, err = wc1.Write([]byte(masterKey + "\n"))
	//time.Sleep(time.Second)
	//out = <-outChan
	//wc1.Close()
	//fmt.Println(out)

	//_, out = tests.ExecuteT(t, "gaiacli keys show foo")
	//fooAddr := strings.TrimLeft(out, "foo\t")

	//wc2, _ := tests.GoExecuteT(t, "gaiad start")
	//defer wc2.Close()
	//time.Sleep(time.Second)

	//_, out = tests.ExecuteT(t, fmt.Sprintf("gaiacli account %v", fooAddr))
	//fmt.Println(fooAddr)
}
