// https://raw.githubusercontent.com/roboll/go-vendorinstall/a3e9f0a5d5861b3bb16b93200b2c359c9846b3c5/main.go

package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	source   = flag.String("source", "vendor", "source directory")
	target   = flag.String("target", "", "target directory (defaults to $GOBIN, if not set $GOPATH/bin)")
	commands = flag.String("commands", "", "comma separated list of commands to execute after go install in temporary environment")
	quiet    = flag.Bool("quiet", false, "disable output")
)

func main() {
	flag.Parse()

	packages := flag.Args()
	if len(packages) < 1 {
		fail(errors.New("no packages: specify a package"))
	}

	gopath, err := ioutil.TempDir("", "go-vendorinstall-gopath")
	if err != nil {
		fail(err)
	}
	print(fmt.Sprintf("gopath: %s", gopath))
	defer func() {
		if err := os.RemoveAll(gopath); err != nil {
			fail(err)
		}
	}()

	if len(*target) == 0 {
		if gobin := os.Getenv("GOBIN"); len(gobin) > 0 {
			target = &gobin
		} else {
			bin := fmt.Sprintf("%s/bin", os.Getenv("GOPATH"))
			target = &bin
		}
	}

	gobin, err := filepath.Abs(*target)
	if err != nil {
		fail(err)
	}
	print(fmt.Sprintf("gobin: %s", gobin))

	if err := link(gopath, *source); err != nil {
		fail(err)
	}

	oldpath := os.Getenv("PATH")
	path := fmt.Sprintf("%s%s%s", gobin, string(os.PathListSeparator), os.Getenv("PATH"))
	os.Setenv("PATH", fmt.Sprintf("%s%s%s", gobin, string(os.PathListSeparator), os.Getenv("PATH")))
	defer os.Setenv("PATH", oldpath)

	env := []string{fmt.Sprintf("PATH=%s", path), fmt.Sprintf("GOPATH=%s", gopath), fmt.Sprintf("GOBIN=%s", gobin)}
	args := append([]string{"install"}, packages...)
	if out, err := doexec("go", gopath, args, env); err != nil {
		print(string(out))
		fail(err)
	}

	if len(*commands) > 0 {
		for _, cmd := range strings.Split(*commands, ",") {
			split := strings.Split(cmd, " ")
			if out, err := doexec(split[0], gopath, split[1:], env); err != nil {
				print(string(out))
				fail(err)
			}
		}
	}
}

func print(msg string) {
	if !*quiet {
		fmt.Println(msg)
	}
}

func fail(err error) {
	fmt.Printf("error: %s", err.Error())
	os.Exit(1)
}

func link(gopath, source string) error {
	srcdir, err := filepath.Abs(source)
	if err != nil {
		return err
	}

	linkto := filepath.Join(gopath, "src")
	if err := os.MkdirAll(linkto, 0777); err != nil {
		return err
	}

	files, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, file := range files {
		real := filepath.Join(srcdir, file.Name())
		link := filepath.Join(linkto, file.Name())
		if err := os.Symlink(real, link); err != nil {
			return err
		}
	}

	return nil
}

func doexec(bin, dir string, args []string, env []string) ([]byte, error) {
	print(fmt.Sprintf("%s %s", bin, strings.Join(args, " ")))
	cmd := exec.Command(bin, args...)
	cmd.Env = env
	cmd.Dir = dir

	return cmd.CombinedOutput()
}
