package cmd

import (
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cosmos/cosmos-sdk/version"
	tmversion "github.com/tendermint/tendermint/version"
	"github.com/spf13/cobra"
	"path/filepath"
)

var remoteProjectPath string

func init() {
	initCmd.Flags().StringVarP(&remoteProjectPath, "project-path", "p", "", "Remote project path. eg: github.com/your_user_name/project_name")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
        Use:   "init [ProjectName]",
        Short: "Initialize your new cosmos zone",
        Args:  cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
                fmt.Print("Thanks for choosing Cosmos-SDK to build your project.\n\n")
                projectName := args[0]
                shortProjectName := strings.ToLower(projectName)
                remoteProjectPath = strings.ToLower(strings.TrimSpace(remoteProjectPath))
                if remoteProjectPath == "" {
                        remoteProjectPath = strings.ToLower(shortProjectName)
                }
                setupBasecoinWorkspace(shortProjectName, remoteProjectPath)
                return nil
        },
}


func resolveProjectPath(remoteProjectPath string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
		// Use $HOME/go
	}
	return gopath + string(os.PathSeparator) + "src" + string(os.PathSeparator) + remoteProjectPath
}

var remoteBasecoinPath = "github.com/cosmos/cosmos-sdk/examples/basecoin"

func copyBasecoinTemplate(projectName string, projectPath string, remoteProjectPath string) {
	basecoinProjectPath := resolveProjectPath(remoteBasecoinPath)
	filepath.Walk(basecoinProjectPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			contents := string(data)
			// Extract relative file path eg: app/app.go instead of /Users/..../github.com/cosmos/...examples/basecoin/app/app.go
			relativeFilePath := path[len(basecoinProjectPath)+1:]
			// Evaluating the filepath in the new project folder
			projectFilePath := projectPath + string(os.PathSeparator) + relativeFilePath
			lengthOfRootDir := strings.LastIndex(projectFilePath, string(os.PathSeparator))
			// Extracting the path of root directory from the filepath
			rootDir := projectFilePath[0:lengthOfRootDir]
			// Creating the required directory first
			os.MkdirAll(rootDir, os.ModePerm)
			fmt.Println("Creating " + projectFilePath)
			// Writing the contents to a file in the project folder
			ioutil.WriteFile(projectFilePath, []byte(contents), os.ModePerm)
		}
		return nil
	})
}

func createGopkg(projectPath string) {
	// Create gopkg.toml file
	dependencies := map[string]string{
		"github.com/cosmos/cosmos-sdk": "=" + version.Version,
		"github.com/stretchr/testify": "=1.2.1",
		"github.com/spf13/cobra": "=0.0.1",
		"github.com/spf13/viper": "=1.0.0",
	}
	overrides := map[string]string{
		"github.com/golang/protobuf": "1.1.0",
		"github.com/tendermint/tendermint": tmversion.Version,
	}
	contents := ""
	for dependency, version := range dependencies {
		contents += "[[constraint]]\n\tname = \"" + dependency + "\"\n\tversion = \"" + version + "\"\n\n"
	}
	for dependency, version := range overrides {
		contents += "[[override]]\n\tname = \"" + dependency + "\"\n\tversion = \"=" + version + "\"\n\n"
	}
	contents += "[prune]\n\tgo-tests = true\n\tunused-packages = true"
	ioutil.WriteFile(projectPath+"/Gopkg.toml", []byte(contents), os.ModePerm)
}

func createMakefile(projectPath string) {
	// Create makefile
	makefileContents := `PACKAGES=$(shell go list ./... | grep -v '/vendor/')

all: get_tools get_vendor_deps build test

get_tools:
	go get github.com/golang/dep/cmd/dep

build:
	go build -o bin/basecli cmd/basecli/main.go && go build -o bin/basecoind cmd/basecoind/main.go

get_vendor_deps:
	@rm -rf vendor/
	@dep ensure

test:
	@go test $(PACKAGES)

benchmark:
	@go test -bench=. $(PACKAGES)

.PHONY: all build test benchmark`
	ioutil.WriteFile(projectPath+"/Makefile", []byte(makefileContents), os.ModePerm)

}

func setupBasecoinWorkspace(projectName string, remoteProjectPath string) {
	projectPath := resolveProjectPath(remoteProjectPath)
	fmt.Println("Configuring your project in " + projectPath)
	copyBasecoinTemplate(projectName, projectPath, remoteProjectPath)
	createGopkg(projectPath)
	createMakefile(projectPath)
	fmt.Printf("\nInitialized a new project at %s.\nHappy hacking!\n", projectPath)
}

