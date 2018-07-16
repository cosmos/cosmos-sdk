package cmd

import (
	"bufio"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/gobuffalo/packr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

func resolveProjectPath(remoteProjectPath string) string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
		// Use $HOME/go
	}
	return gopath + string(os.PathSeparator) + "src" + string(os.PathSeparator) + remoteProjectPath
}

var initCmd = &cobra.Command{
	Use:   "init AwesomeProjectName",
	Short: "Initialize your new cosmos zone",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Project name is required")
		}
		projectName := args[0]
		capitalizedProjectName := strings.Title(projectName)
		shortProjectName := strings.ToLower(projectName)
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("Thank you for using cosmos-zone tool.")
		fmt.Println("You are only a few steps away from creating your brand new blockchain project on Cosmos.")
		fmt.Print("We will ask you a few more questions to guide you through this process\n\n")
		fmt.Print("To configure this project we need a remote project path. If you are unsure you can leave this empty. ")
		fmt.Print("Remote project path is usually something like github.com/your_user_name/project_name\n")
		fmt.Print("Enter remote project path: ")
		remoteProjectPath, _ := reader.ReadString('\n')
		remoteProjectPath = strings.ToLower(strings.TrimSpace(remoteProjectPath))
		if remoteProjectPath == "" {
			remoteProjectPath = strings.ToLower(shortProjectName)
		}
		projectPath := resolveProjectPath(remoteProjectPath)
		fmt.Print("configuring the project in " + projectPath + "\n\n")
		time.Sleep(2 * time.Second)
		box := packr.NewBox("../template")
		var replacer = strings.NewReplacer("_CAPITALIZED_PROJECT_SHORT_NAME_", capitalizedProjectName, "_PROJECT_SHORT_NAME_", shortProjectName, "_REMOTE_PROJECT_PATH_", remoteProjectPath)
		box.Walk(func(path string, file packr.File) error {
			actualPath := replacer.Replace(path)
			fmt.Println("Creating file: " + actualPath)
			contents := box.String(path)
			contents = replacer.Replace(contents)
			lastIndex := strings.LastIndex(actualPath, string(os.PathSeparator))
			rootDir := ""
			if lastIndex != -1 {
				rootDir = actualPath[0:lastIndex]
			}
			// Create directory
			os.MkdirAll(projectPath+string(os.PathSeparator)+rootDir, os.ModePerm)
			filePath := projectPath + string(os.PathSeparator) + actualPath
			ioutil.WriteFile(filePath, []byte(contents), os.ModePerm)
			return nil
		})
		fmt.Println("Initialized a new project at " + projectPath + ". Happy hacking!")
		return nil
	},
}
