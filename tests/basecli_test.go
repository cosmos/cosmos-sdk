package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests assume the `basecoind` and `basecli` binaries
// have been built and are located in `./build`

var (
	gopath = filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "cosmos", "cosmos-sdk")

	basecoind = "build/basecoind"
	basecli   = "build/basecli"

	basecoindPath = filepath.Join(gopath, basecoind)
	basecliPath   = filepath.Join(gopath, basecli)

	basecoindDir = "./tmp-basecoind-tests"
	basecliDir   = "./tmp-basecli-tests"

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

func TestInitBasecoin(t *testing.T) {
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

	// left here as a sanity test
	makeKeys(t)
}

// identical to above test but doesn't make keys
// and returns the address as string
func initBasecoinServer(t *testing.T) string {
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

// TODO see https://github.com/cosmos/cosmos-sdk/issues/674
func _TestSendCoins(t *testing.T) {
	var err error

	validatorAddress := initBasecoinServer(t)

	err = startServer()
	assert.Nil(t, err)

	// send some coins
	sendTo := fmt.Sprintf("--to=%s", bob)
	sendFrom := fmt.Sprintf("--from=%s", validatorAddress)

	cmdOut, err := exec.Command(basecliPath, "send", sendTo, "--amount=1000mycoin", sendFrom, "--seq=0").Output()
	assert.Nil(t, err)

	fmt.Printf("sent: %s", string(cmdOut))

}

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

// these are in the original bash tests
func TestBaseCliRecover(t *testing.T) {}
func TestBaseCliShow(t *testing.T)    {}

// expects initBasecoinServer to have been run
func startServer(t *testing.T) {
	// straight outta https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
	cmdName := basecoindPath
	cmdArgs := []string{"start", "--home", basecoindDir}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	assert.Nil(t, err)

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("running [basecoind start] %s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	assert.Nil(t, err)

	err = cmd.Wait()
	assert.Nil(t, err)

	time.Sleep(5 * time.Second)

	// TODO return cmd.Process so that we can later do something like:
	// cmd.Process.Kill()
	// see: https://stackoverflow.com/questions/11886531/terminating-a-process-started-with-os-exec-in-golang
}

func cleanUp() {
	// ignore errors b/c the dirs may not yet exist
	os.RemoveAll(basecoindDir)
	os.RemoveAll(basecliDir)
}
