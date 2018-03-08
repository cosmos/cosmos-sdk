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
			fmt.Printf("basecoind start | %s\n", scanner.Text())
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

	return nil
}

func clean() {
	// ignore errors b/c the dirs may not yet exist
	os.Remove(basecoindDir)
	os.Remove(basecliDir)
}
