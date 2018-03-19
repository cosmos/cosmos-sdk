package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	//"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Tests assume the `basecoind` and `basecli` binaries
// have been built and are located in `./build`

// TODO remove test dirs if tests are successful

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
)

func TestMain(m *testing.M) {
	cleanUp()

	m.Run()

	cleanUp()
}

func TestInitBaseCoin(t *testing.T) {

	var err error

	password := "some-random-password"
	usePassword := exec.Command("echo", password)

	initBasecoind := exec.Command(basecoindPath, "init", "--home", basecoindDir)

	initBasecoind.Stdin, err = usePassword.StdoutPipe()
	if err != nil {
		t.Error(err)
	}

	initBasecoind.Stdout = os.Stdout

	if err := initBasecoind.Start(); err != nil {
		t.Error(err)
	}
	if err := usePassword.Run(); err != nil {
		t.Error(err)
	}
	if err := initBasecoind.Wait(); err != nil {
		t.Error(err)
	}

	if err := makeKeys(); err != nil {
		t.Error(err)
	}
}

func makeKeys() error {
	var err error
	for _, acc := range ACCOUNTS {
		pass := exec.Command("echo", "1234567890")
		makeKeys := exec.Command(basecliPath, "keys", "add", acc, "--home", basecliDir)

		makeKeys.Stdin, err = pass.StdoutPipe()
		if err != nil {
			return err
		}

		makeKeys.Stdout = os.Stdout
		if err := makeKeys.Start(); err != nil {
			return err
		}
		if err := pass.Run(); err != nil {
			return err
		}

		if err := makeKeys.Wait(); err != nil {
			return err
		}
	}

	return nil
}

// these are in the original bash tests
func TestBaseCliRecover(t *testing.T) {}
func TestBaseCliShow(t *testing.T)    {}

func _TestSendCoins(t *testing.T) {
	if err := startServer(); err != nil {
		t.Error(err)
	}

	// send some coins
	// [zr] where dafuq do I get a FROM (oh, use --name)

	sendTo := fmt.Sprintf("--to=%s", bob)
	sendFrom := fmt.Sprintf("--from=%s", alice)

	cmdOut, err := exec.Command(basecliPath, "send", sendTo, "--amount=1000mycoin", sendFrom, "--seq=0").Output()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("sent: %s", string(cmdOut))

}

// expects TestInitBaseCoin to have been run
func startServer() error {
	// straight outta https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
	cmdName := basecoindPath
	cmdArgs := []string{"start", "--home", basecoindDir}

	cmd := exec.Command(cmdName, cmdArgs...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("running [basecoind start] %s\n", scanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	time.Sleep(5 * time.Second)

	return nil

	// TODO return cmd.Process so that we can later do something like:
	// cmd.Process.Kill()
	// see: https://stackoverflow.com/questions/11886531/terminating-a-process-started-with-os-exec-in-golang
}

func cleanUp() {
	// ignore errors b/c the dirs may not yet exist
	os.RemoveAll(basecoindDir)
	os.RemoveAll(basecliDir)
}
