package lcd

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/keys"
)

func main() {
	kb, err := keys.NewKeyBaseFromDir(InitClientHome(""))
	if err != nil {
		panic(err)
	}
	addr, _, err := CreateAddr("contract_tester", "contract_tester", kb)
	cleanup, valConsPubKeys, valOperAddrs, port := InitializeTestLCD(t, 1, []sdk.AccAddress{addr}, true, "58645")
	defer cleanup()
	fmt.Println("TEST", valConsPubKeys, port, valOperAddrs)
}
