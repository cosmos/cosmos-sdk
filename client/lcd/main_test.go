package lcd

import (
	"fmt"
	"os"
	"testing"

	nm "github.com/tendermint/tendermint/node"
)

var node *nm.Node

// See https://golang.org/pkg/testing/#hdr-Main
// for more details
func TestMain(m *testing.M) {
	// start a basecoind node and LCD server in the background to test against

	// run all the tests against a single server instance
	node, lcd, err := startTMAndLCD()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	code := m.Run()

	// tear down
	// TODO: cleanup
	// TODO: it would be great if TM could run without
	// persiting anything in the first place
	node.Stop()
	node.Wait()

	// just a listener ...
	lcd.Close()

	os.Exit(code)
}
