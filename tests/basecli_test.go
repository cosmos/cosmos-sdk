package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	//"strings"
	"testing"
)

// Tests assume the `basecoind` and `basecli` binaries
// have been built and are located in `./build`

var (
	basecoind = "build/basecoind"
	basecli   = "build/basecli"

	basecoindDir = "./basecoind-tests"
	basecliDir   = "./basecli-tests"

	from = "demo" // but we need to create the named key first ... ?
	to   = "ABCAFE00DEADBEEF00CAFE00DEADBEEF00CAFE00"
)

func gopath() string {
	return filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "cosmos", "cosmos-sdk")
}

func whereisBasecoind() string {
	return filepath.Join(gopath(), basecoind)
}

func whereisBasecli() string {
	return filepath.Join(gopath(), basecli)
}

func TestInitBaseCoin(t *testing.T) {
	clean()

	var err error

	password := "some-random-password"
	usePassword := exec.Command("echo", password)

	initBasecoind := exec.Command(whereisBasecoind(), "init", "--home", basecoindDir)

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
}

// these are in the original bash tests
func TestBaseCliRecover(t *testing.T) {}
func TestBaseCliShow(t *testing.T)    {}

func TestSendCoins(t *testing.T) {
	if err := startServer(); err != nil {
		t.Error(err)
	}

	// send some coins
	// [zr] where dafuq do I get a FROM (oh, use --name)

	sendTo := fmt.Sprintf("--to=%s", to)
	sendFrom := fmt.Sprintf("--from=%s", from)

	cmdOut, err := exec.Command(whereisBasecli(), "send", sendTo, "--amount=1000mycoin", sendFrom, "--seq=0")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("sent: %s", string(cmdOut))

}

// expects TestInitBaseCoin to have been run
func startServer() error {
	// straight outta https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
	cmdName := whereisBasecoind()
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

	time.Sleep(5 * time.Seconds())

	return nil

	// TODO return cmd.Process so that we can later do something like:
	// cmd.Process.Kill()
	// see: https://stackoverflow.com/questions/11886531/terminating-a-process-started-with-os-exec-in-golang
}

func clean() {
	// ignore errors b/c the dirs may not yet exist
	os.Remove(basecoindDir)
	os.Remove(basecliDir)
}

/*

initial attempt from gaia repo
var (
	GAIA       = "gaia"
	SERVER_EXE = "node"
	CLIENT_EXE = "client"
	ACCOUNTS   = []string{"alice", "bob", "charlie", "igor"}
	RICH       = ACCOUNTS[0]
	CANDIDATE  = ACCOUNTS[1]
	DELEGATOR  = ACCOUNTS[2]
	POOR       = ACCOUNTS[3]

	chainID = "staking_test"
	testDir = "./tmp_tests"
)

func runTests() {

	if err := os.Mkdir(testDir, 0666); err != nil {
		panic(err)
	}
	defer os.Remove(testDir)

	// make some keys

	//if err := makeKeys(); err != nil {
	//	panic(err)
	//}

	if err := initServer(); err != nil {
		fmt.Printf("Err: %v", err)
		panic(err)
	}

}

func makeKeys() error {
	fmt.Println("make keys")
	var err error
	for _, acc := range ACCOUNTS {
		pass := exec.Command("echo", "1234567890")
		makeKeys := exec.Command(GAIA, CLIENT_EXE, "keys", "new", acc)

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

		fmt.Printf("OUT: %v", makeKeys)
	}

	return nil
}

func initServer() error {
	serveDir := filepath.Join(testDir, "server")
	//serverLog := filepath.Join(testDir, "gaia-node.log")

	// get RICH
	keyOut, err := exec.Command(GAIA, CLIENT_EXE, "keys", "get", "alice").Output()
	if err != nil {
		fmt.Println("one")
		return err
	}
	key := strings.Split(string(keyOut), "\t")
	fmt.Printf("wit:%s", key[2])

	outByte, err := exec.Command(GAIA, SERVER_EXE, "init", "--static", fmt.Sprintf("--chain-id=%s", chainID), fmt.Sprintf("--home=%s", serveDir), key[2]).Output()
	if err != nil {
		fmt.Println("teo")
		fmt.Printf("Error: %v", err)

		return err
	}
	fmt.Sprintf("OUT: %s", string(outByte))
	return nil
}

*/
