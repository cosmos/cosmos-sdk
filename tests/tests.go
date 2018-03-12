package tests

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	//"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Tests assume the `basecoind` and `basecli` binaries
// have been built and are located in `./build`

// TODO remove test dirs if tests are successful

//nolint
var (
	basecoind = "build/basecoind"
	basecli   = "build/basecli"

	basecoindDir = "./tmp-basecoind-tests"
	basecliDir   = "./tmp-basecli-tests"

	ACCOUNTS = []string{"alice", "bob", "charlie", "igor"}
	alice    = ACCOUNTS[0]
	bob      = ACCOUNTS[1]
	charlie  = ACCOUNTS[2]
	igor     = ACCOUNTS[3]
)

func gopath() string {
	return filepath.Join(os.Getenv("GOPATH"), "src", "github.com", "cosmos", "cosmos-sdk")
}

func whereIsBasecoind() string {
	return filepath.Join(gopath(), basecoind)
}

func whereIsBasecli() string {
	return filepath.Join(gopath(), basecli)
}

// Init Basecoin Test
func TestInitBasecoin(t *testing.T) {
	Clean()

	var err error

	password := "some-random-password"
	usePassword := exec.Command("echo", password)

	initBasecoind := exec.Command(whereIsBasecoind(), "init", "--home", basecoindDir)

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
		makeKeys := exec.Command(whereIsBasecli(), "keys", "add", acc, "--home", basecliDir)

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

func _TestSendCoins(t *testing.T) {
	if err := StartServer(); err != nil {
		t.Error(err)
	}

	// send some coins
	// [zr] where dafuq do I get a FROM (oh, use --name)

	sendTo := fmt.Sprintf("--to=%s", bob)
	sendFrom := fmt.Sprintf("--from=%s", alice)

	cmdOut, err := exec.Command(whereIsBasecli(), "send", sendTo, "--amount=1000mycoin", sendFrom, "--seq=0").Output()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("sent: %s", string(cmdOut))

}

// expects TestInitBaseCoin to have been run
func StartServer() error {
	// straight outta https://nathanleclaire.com/blog/2014/12/29/shelled-out-commands-in-golang/
	cmdName := whereIsBasecoind()
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

// expects TestInitBaseCoin to have been run
func StartServerForTest(t *testing.T) *exec.Cmd {
	cmdName := whereIsBasecoind()
	cmdArgs := []string{"start", "--home", basecoindDir}
	cmd := exec.Command(cmdName, cmdArgs...)
	err := cmd.Start()
	require.Nil(t, err)
	return cmd
}

// clean the directories
func Clean() {
	// ignore errors b/c the dirs may not yet exist
	os.Remove(basecoindDir)
	os.Remove(basecliDir)
}

/*

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
