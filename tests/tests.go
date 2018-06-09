package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
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
func TestInitBasecoin(t *testing.T, home string) string {
	var err error

	password := "some-random-password"

	initBasecoind := exec.Command(whereIsBasecoind(), "init", "--home", home)
	cmdWriter, err := initBasecoind.StdinPipe()
	require.Nil(t, err)

	buf := new(bytes.Buffer)
	initBasecoind.Stdout = buf

	if err = initBasecoind.Start(); err != nil {
		t.Error(err)
	}

	_, err = cmdWriter.Write([]byte(password))
	require.Nil(t, err)
	cmdWriter.Close()

	if err = initBasecoind.Wait(); err != nil {
		t.Error(err)
	}

	// get seed from initialization
	theOutput := strings.Split(buf.String(), "\n")
	var seedLine int
	for _seedLine, o := range theOutput {
		if strings.HasPrefix(string(o), "Secret phrase") {
			seedLine = _seedLine + 1
			break
		}
	}

	seed := string(theOutput[seedLine])

	// enable indexing
	err = appendToFile(path.Join(home, "config", "config.toml"), "\n\n[tx_indexing]\nindex_all_tags = true\n")
	require.Nil(t, err)

	return seed
}

func appendToFile(path string, text string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err = f.WriteString(text); err != nil {
		return err
	}

	return nil
}

func makeKeys() error {
	for _, acc := range ACCOUNTS {
		makeKeys := exec.Command(whereIsBasecli(), "keys", "add", acc, "--home", basecliDir)
		cmdWriter, err := makeKeys.StdinPipe()
		if err != nil {
			return err
		}

		makeKeys.Stdout = os.Stdout
		if err := makeKeys.Start(); err != nil {
			return err
		}
		cmdWriter.Write([]byte("1234567890"))
		if err != nil {
			return err
		}
		cmdWriter.Close()

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

// Init Basecoin Test
func InitServerForTest(t *testing.T) {
	Clean()

	var err error

	password := "some-random-password"
	usePassword := exec.Command("echo", password)

	initBasecoind := exec.Command(whereIsBasecoind(), "init", "--home", basecoindDir)

	initBasecoind.Stdin, err = usePassword.StdoutPipe()
	require.Nil(t, err)

	initBasecoind.Stdout = os.Stdout

	err = initBasecoind.Start()
	require.Nil(t, err)
	err = usePassword.Run()
	require.Nil(t, err)
	err = initBasecoind.Wait()
	require.Nil(t, err)

	err = makeKeys()
	require.Nil(t, err)
}

// expects TestInitBaseCoin to have been run
func StartNodeServerForTest(t *testing.T, home string) *exec.Cmd {
	cmdName := whereIsBasecoind()
	cmdArgs := []string{"start", "--home", home}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	require.Nil(t, err)

	// FIXME: if there is a nondeterministic node start failure,
	//        we should probably make this read the logs to wait for RPC
	time.Sleep(time.Second * 2)

	return cmd
}

// expects TestInitBaseCoin to have been run
func StartLCDServerForTest(t *testing.T, home, chainID string) (cmd *exec.Cmd, port string) {
	cmdName := whereIsBasecli()
	var err error
	_, port, err = server.FreeTCPAddr()
	require.NoError(t, err)
	cmdArgs := []string{
		"rest-server",
		"--home",
		home,
		"--bind",
		fmt.Sprintf("localhost:%s", port),
		"--chain-id",
		chainID,
	}
	cmd = exec.Command(cmdName, cmdArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	require.Nil(t, err)
	time.Sleep(time.Second * 2) // TODO: LOL
	return cmd, port
}

// clean the directories
func Clean() {
	// ignore errors b/c the dirs may not yet exist
	err := os.Remove(basecoindDir)
	panic(err)
	err = os.Remove(basecliDir)
	panic(err)
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
