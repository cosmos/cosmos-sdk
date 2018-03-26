package main

import (
	//"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/*

Cosmos-SDK integration tests using golang (instead of bash)

Tests assume the `basecoind` and `basecli` binaries
have been built and are located in `./build`
The CI handles this already

NOTE: a few Test* functions have duplicate functions
that aren't specifically "tests" but rather pre-requisite
functionality that's called from within a particular Test*
function. We do this because Test* functions provide reporting
capabilities for the CI. This duplication is considered an
acceptable trade-off.

*/

var (
	gopath = filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "cosmos", "cosmos-sdk")

	basecoind = "build/basecoind"
	basecli   = "build/basecli"

	basecoindPath = filepath.Join(gopath, basecoind)
	basecliPath   = filepath.Join(gopath, basecli)

	basecoindDir    = "./tmp-basecoind-tests"
	basecliDir      = "./tmp-basecli-tests"
	files_for_tests = "./files-for-tests"

	ACCOUNTS = []string{"alice", "bob", "charlie", "igor"}
	alice    = ACCOUNTS[0]
	bob      = ACCOUNTS[1]
	charlie  = ACCOUNTS[2]
	igor     = ACCOUNTS[3]

	password = "some-random-password-doesnt-matter"
)

func TestMain(m *testing.M) {
	// start by cleaning up because test dirs get left
	// behind if tests are manually exited
	cleanUp()

	m.Run()

	cleanUp()
}

// `basecoind init`
func _TestInitBasecoin(t *testing.T) {
	var err error

	usePassword := exec.Command("echo", password)

	initBasecoind := exec.Command(basecoindPath, "init", "--home", basecoindDir)

	initBasecoind.Stdin, err = usePassword.StdoutPipe()
	assert.Nil(t, err)

	initBasecoind.Stdout = os.Stdout

	err = initBasecoind.Start()
	assert.Nil(t, err)

	err = usePassword.Run()
	assert.Nil(t, err)

	err = initBasecoind.Wait()
	assert.Nil(t, err)
}

// identical to above test but doesn't make keys
// and returns the address as string
func initBasecoindServer(t *testing.T) string {
	var err error

	usePassword := exec.Command("echo", password)

	initBasecoind := exec.Command(basecoindPath, "init", "--home", basecoindDir)

	initBasecoind.Stdin, err = usePassword.StdoutPipe()
	assert.Nil(t, err)

	initBasecoind.Stdout = os.Stdout

	err = initBasecoind.Start()
	assert.Nil(t, err)

	err = usePassword.Run()
	assert.Nil(t, err)

	err = initBasecoind.Wait()
	assert.Nil(t, err)

	// TODO get the address!

	return ""

}

// `basecli keys add`
// duplicate of TestBasecliKeysAdd
// but see note at top of file
func makeKeys(t *testing.T) {
	var err error
	for _, acc := range ACCOUNTS {
		pass := exec.Command("echo", password)
		makeKeys := exec.Command(basecliPath, "keys", "add", acc, "--home", basecliDir)

		makeKeys.Stdin, err = pass.StdoutPipe()
		assert.Nil(t, err)

		makeKeys.Stdout = os.Stdout
		err = makeKeys.Start()
		assert.Nil(t, err)

		err = pass.Run()
		assert.Nil(t, err)

		err = makeKeys.Wait()
		assert.Nil(t, err)
	}
}

// `basecli init`
func TestBasecliInit(t *testing.T) {}

// `basecli rest-server`
func TestBasecliRestServer(t *testing.T) {}

// `basecli status`
func TestBasecliStatus(t *testing.T) {}

// `basecli block`
func TestBasecliBlock(t *testing.T) {}

// `basecli validatorset`
func TestBasecliValidatorSet(t *testing.T) {}

// `basecli txs`
func TestBasecliTxs(t *testing.T) {}

// `basecli tx`
func TestBasecliTx(t *testing.T) {}

// `basecli account`
func TestBasecliAccount(t *testing.T) {}

// TODO see https://github.com/cosmos/cosmos-sdk/issues/674
// instead of the above, we're using already init'd files
// `basecli send`
func TestBasecliSend(t *testing.T) {
	var err error

	// seed that created the file priv_validator.json in tests/files-for-tests/config/priv_validator.json
	seed := "\"choose method diagram error travel conduct juice loop calm ridge gesture reason damp spider arm abandon\""
	namedKey := "badux"
	// address from that seed
	//validatorAddress := "7AA8E48D709A9E28EA9E81026522F4A924FDA3E7"

	// make a named key (TODO, remove this awful UX)
	err = addNamedKeyFromSeed(namedKey, seed)
	assert.Nil(t, err)

	// copy the pre-made bob key to the temp dir
	err = os.Link(filepath.Join(files_for_tests, "keys.db"), filepath.Join(basecliPath, "keys.db"))
	assert.Nil(t, err)

	// start the server
	toKill := startServer(t)

	// send some coins
	sendTo := fmt.Sprintf("--to=%s", bob)
	sendFrom := fmt.Sprintf("--name=%s", namedKey)

	cmdOut, err := exec.Command(basecliPath, "send", sendTo, "--amount=1000mycoin", sendFrom, "--seq=0", "--home", basecliPath).Output()
	assert.Nil(t, err)

	fmt.Printf("sent: %s", string(cmdOut))

	// kill startServer
	// see: https://stackoverflow.com/questions/11886531/terminating-a-process-started-with-os-exec-in-golang
	toKill.Process.Kill()

}

func addNamedKeyFromSeed(name, seed string) error {
	var err error

	useSeed := exec.Command("echo", password, ";", "echo", seed)

	addKey := exec.Command(basecliPath, "keys", "add", name, "--home", basecliPath, "--recover")

	addKey.Stdin, err = useSeed.StdoutPipe()
	if err != nil {
		return err
	}

	addKey.Stdout = os.Stdout

	err = addKey.Start()
	if err != nil {
		return err
	}

	err = useSeed.Run()
	if err != nil {
		return err
	}

	err = addKey.Wait()
	if err != nil {
		return err
	}

	out, err := addKey.Output()
	if err != nil {
		return err
	}

	fmt.Printf("name (%s) added: %s", name, string(out))
	return nil
}

// `basecli transfer`
func TestBasecliTransfer(t *testing.T) {}

// `basecli relay`
func TestBasecliRelay(t *testing.T) {}

// `basecli declare-candidacy`
func TestBasecliDeclareCandidacy(t *testing.T) {}

// `basecli bond`
func TestBasecliBond(t *testing.T) {}

// `basecli unbond`
func TestBasecliUnbond(t *testing.T) {}

// `basecli keys add`
// duplicate with "makeKeys" function
func _TestBasecliKeysAdd(t *testing.T) {
	_ = initBasecoindServer(t)

	var err error
	for _, acc := range ACCOUNTS {
		pass := exec.Command("echo", password)
		makeKeys := exec.Command(basecliPath, "keys", "add", acc, "--home", basecliDir)

		makeKeys.Stdin, err = pass.StdoutPipe()
		assert.Nil(t, err)

		makeKeys.Stdout = os.Stdout
		err = makeKeys.Start()
		assert.Nil(t, err)

		err = pass.Run()
		assert.Nil(t, err)

		err = makeKeys.Wait()
		assert.Nil(t, err)
	}

}

// `basecli keys add --recover`
func TestBasecliKeysAddRecover(t *testing.T) {}

// `basecli keys list`
func TestBasecliKeysList(t *testing.T) {}

// `basecli keys show`
func TestBasecliKeysShow(t *testing.T) {}

// `basecli keys delete`
func TestBasecliKeysDelete(t *testing.T) {}

// `basecli keys update`
func TestBasecliKeysUpdate(t *testing.T) {}

// `basecoind start`
func startServer(t *testing.T) *exec.Cmd {
	var err error

	cmdArgs := []string{"start", "--home", basecoindDir}

	cmd := exec.Command(basecoindPath, cmdArgs...)
	//cmd.Stdout = os.Stdout

	err = cmd.Start()
	assert.Nil(t, err)

	time.Sleep(2 * time.Second)

	return cmd

}

func cleanUp() {
	// ignore errors b/c the dirs may not yet exist
	os.RemoveAll(basecoindDir)
	os.RemoveAll(basecliDir)
}
