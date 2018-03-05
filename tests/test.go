package main

import (
	//	"flag"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {

	switch os.Args[1] {
	case "test":
		runTests()
	default:
		fmt.Println("command not yet implemented")
	}

}

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

/*func RunCommand(cmd *exec.Cmd) {
	fmt.Println(">", strings.Join(cmd.Args, " "))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}

func RunCommandWithArgs(cmd string, args ...string) {
	RunCommand(exec.Command(cmd, args...))
}*/

// Execute - Execute the command, return standard output and error
func Execute(command string) (stdOut string, err error) {
	//split command into command and args
	var outByte []byte
	split := strings.Split(command, " ")
	switch len(split) {
	case 0:
		return "", errors.New("no command provided")
	case 1:
		outByte, err = exec.Command(split[0]).Output()
	default:
		outByte, err = exec.Command(split[0], split[1:]...).Output()
	}
	stdOut = string(outByte)
	stdOut = strings.Trim(stdOut, "\n") //trim any new lines
	return
}
