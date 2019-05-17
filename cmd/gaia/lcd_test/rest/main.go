package main

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/lcd_test"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	kb, err := keys.NewKeyBaseFromDir(lcdtest.InitClientHome(""))
	if err != nil {
		panic(err)
	}
	addr, p, err := lcdtest.CreateAddr("contract_tester", "contract_tester", kb)
	if err != nil {
		panic(err)
	}
	// 85B0FC5010CBAEBB58C6914DFF890982F7404374 pause fun stairs ready amount radar travel wrist present guitar awake stand speed leg local giant taxi crime dirt arrange rifle width avocado virtual
	fmt.Println(addr, p)
	cleanup, valConsPubKeys, valOperAddrs, port := lcdtest.InitializeLCD(1, []sdk.AccAddress{addr}, true, "58645")
	fmt.Println("TEST", valConsPubKeys, port, valOperAddrs)
	defer cleanup()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, os.Interrupt, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGHUP)
	go func() {
		sig := <-sigs
		fmt.Println("Received", sig)
		done <- true
	}()

	fmt.Println("REST server running")
	<-done
	fmt.Println("exiting")
}
