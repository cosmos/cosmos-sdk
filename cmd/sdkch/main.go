package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

const (
	entriesDirName         = ".pending"
	ghLinkPattern          = `#([0-9]+)`
	ghLinkExpanded         = `[\#$1](https://github.com/cosmos/cosmos-sdk/issues/$1)`
	maxEntryFilenameLength = 20
)

var (
	progName   string
	verboseLog *log.Logger

	entriesDir      string
	verboseLogging  bool
	interactiveMode bool

	// sections name-title map
	sections = map[string]string{
		"breaking":     "Breaking Changes",
		"features":     "New features",
		"improvements": "Improvements",
		"bugfixes":     "Bugfixes",
	}
	// stanzas name-title map
	stanzas = map[string]string{
		"gaia":       "Gaia",
		"gaiacli":    "Gaia CLI",
		"gaiarest":   "Gaia REST API",
		"sdk":        "SDK",
		"tendermint": "Tendermint",
	}

	// RootCmd represents the base command when called without any subcommands
	RootCmd = &cobra.Command{
		Use:   "sdkch",
		Short: "\nMaintain unreleased changelog entries in a modular fashion.",
	}

	// command to add a pending log entry
	AddCmd = &cobra.Command{
		Use:   "add [section] [stanza] [message]",
		Short: "Add an entry file.",
		Long: `
Add an entry file. If message is empty, start the editor to edit the message.

    Sections             Stanzas
         ---                 ---
    breaking                gaia
    features             gaiacli
improvements            gaiarest
    bugfixes                 sdk
                      tendermint`,
		Args: cobra.MaximumNArgs(3),
		Run: func(cmd *cobra.Command, args []string) {

			if interactiveMode {
				addEntryFileFromConsoleInput()
				return
			}

			if len(args) < 2 {
				log.Println("must include at least 2 arguments when not in interactive mode")
				return
			}
			sectionDir, stanzaDir := args[0], args[1]
			validateSectionStanzaDirs(sectionDir, stanzaDir)
			if len(args) == 3 {
				addSinglelineEntryFile(sectionDir, stanzaDir, strings.TrimSpace(args[2]))
				return
			}
			addEntryFile(sectionDir, stanzaDir)
		},
	}

	// command to generate the changelog
	GenerateCmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a changelog in Markdown format and print it to STDOUT. version defaults to UNRELEASED.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			version := "UNRELEASED"
			if flag.NArg() > 1 {
				version = strings.Join(flag.Args()[1:], " ")
			}
			generateChangelog(version)
		},
	}

	// command to delete empty sub-directories recursively
	PruneCmd = &cobra.Command{
		Use:   "prune",
		Short: "Delete empty sub-directories recursively.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			pruneEmptyDirectories()
		},
	}
)

func init() {
	RootCmd.AddCommand(AddCmd)
	RootCmd.AddCommand(GenerateCmd)
	RootCmd.AddCommand(PruneCmd)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	AddCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "get the section/stanza/message with interactive CLI prompts")
	RootCmd.PersistentFlags().BoolVarP(&verboseLogging, "verbose-logging", "v", false, "enable verbose logging")
	RootCmd.PersistentFlags().StringVarP(&entriesDir, "entries-dir", "d", filepath.Join(cwd, entriesDirName), "entry files directory")
}

func main() {

	logPrefix := fmt.Sprintf("%s: ", filepath.Base(progName))
	log.SetFlags(0)
	log.SetPrefix(logPrefix)
	flag.Parse()
	verboseLog = log.New(ioutil.Discard, "", 0)
	if verboseLogging {
		verboseLog.SetOutput(os.Stderr)
		verboseLog.SetPrefix(logPrefix)
	}

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func addEntryFileFromConsoleInput() {

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Please enter the section (either: \"breaking\", \"features\", \"improvements\", \"bugfixes\")")
	sectionDir, _ := reader.ReadString('\n')
	sectionDir = strings.TrimSpace(sectionDir)
	if _, ok := sections[sectionDir]; !ok {
		log.Fatalln("invalid section, please try again")
	}

	fmt.Println("Please enter the stanza (either: \"gaia\", \"gaiacli\", \"gaiarest\", \"sdk\", \"tendermint\")")
	stanzaDir, _ := reader.ReadString('\n')
	stanzaDir = strings.TrimSpace(stanzaDir)
	if _, ok := stanzas[stanzaDir]; !ok {
		log.Fatalln("invalid stanza, please try again")
	}

	fmt.Println("Please enter the changelog message (or press enter to write in  default $EDITOR)")
	message, _ := reader.ReadString('\n')
	message = strings.TrimSpace(message)
	if message == "" {
		addEntryFile(sectionDir, stanzaDir)
		return
	}

	addSinglelineEntryFile(sectionDir, stanzaDir, message)
}

func addSinglelineEntryFile(sectionDir, stanzaDir, message string) {
	filename := filepath.Join(
		filepath.Join(entriesDir, sectionDir, stanzaDir),
		generateFileName(message),
	)

	writeEntryFile(filename, []byte(message))
}

func addEntryFile(sectionDir, stanzaDir string) {
	bs := readUserInputFromEditor()
	firstLine := strings.TrimSpace(strings.Split(string(bs), "\n")[0])
	filename := filepath.Join(
		filepath.Join(entriesDir, sectionDir, stanzaDir),
		generateFileName(firstLine),
	)

	writeEntryFile(filename, bs)
}

func generateFileName(line string) string {
	var chunks []string

	filenameInvalidChars := regexp.MustCompile(`[^a-zA-Z0-9-_]`)
	subsWithInvalidCharsRemoved := strings.Split(filenameInvalidChars.ReplaceAllString(line, " "), " ")
	for _, sub := range subsWithInvalidCharsRemoved {
		sub = strings.TrimSpace(sub)
		if len(sub) != 0 {
			chunks = append(chunks, sub)
		}
	}

	ret := strings.Join(chunks, "-")
	return ret[:int(math.Min(float64(len(ret)), float64(maxEntryFilenameLength)))]
}

func directoryContents(dirPath string) []os.FileInfo {
	contents, err := ioutil.ReadDir(dirPath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("couldn't read directory %s: %v", dirPath, err)
	}

	if len(contents) == 0 {
		return nil
	}

	// Filter out hidden files
	newContents := contents[:0]
	for _, f := range contents {
		if f.Name()[0] != '.' { // skip hidden files
			newContents = append(newContents, f)
		}
	}
	for i := len(newContents); i < len(contents); i++ {
		contents[i] = nil
	}

	return newContents
}

func generateChangelog(version string) {
	fmt.Printf("# %s\n\n", version)
	for sectionDir, sectionTitle := range sections {
		sectionTitlePrinted := false
		for stanzaDir, stanzaTitle := range stanzas {
			path := filepath.Join(entriesDir, sectionDir, stanzaDir)
			files := directoryContents(path)
			if len(files) == 0 {
				continue
			}

			if !sectionTitlePrinted {
				fmt.Printf("## %s\n\n", sectionTitle)
				sectionTitlePrinted = true
			}

			fmt.Printf("### %s\n\n", stanzaTitle)
			for _, f := range files {
				verboseLog.Println("processing", f.Name())
				filename := filepath.Join(path, f.Name())
				if err := indentAndPrintFile(filename); err != nil {
					log.Fatal(err)
				}
			}

			fmt.Println()
		}
	}
}

func pruneEmptyDirectories() {
	for sectionDir := range sections {
		for stanzaDir := range stanzas {
			mustPruneDirIfEmpty(filepath.Join(entriesDir, sectionDir, stanzaDir))
		}
		mustPruneDirIfEmpty(filepath.Join(entriesDir, sectionDir))
	}
}

// nolint: errcheck
func indentAndPrintFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	firstLine := true
	ghLinkRe := regexp.MustCompile(ghLinkPattern)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 {
			continue
		}

		linkified := ghLinkRe.ReplaceAllString(line, ghLinkExpanded)
		if firstLine {
			fmt.Printf("* %s\n", linkified)
			firstLine = false
			continue
		}

		fmt.Printf("  %s\n", linkified)
	}

	return scanner.Err()
}

// nolint: errcheck
func writeEntryFile(filename string, bs []byte) {
	if err := os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
		log.Fatal(err)
	}
	outFile, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	if _, err := outFile.Write(bs); err != nil {
		log.Fatal(err)
	}

	log.Printf("Unreleased changelog entry written to: %s\n", filename)
	log.Println("To modify this entry please edit or delete the above file directly.")
}

func validateSectionStanzaDirs(sectionDir, stanzaDir string) {
	if _, ok := sections[sectionDir]; !ok {
		log.Fatalf("invalid section -- %s", sectionDir)
	}
	if _, ok := stanzas[stanzaDir]; !ok {
		log.Fatalf("invalid stanza -- %s", stanzaDir)
	}
}

// nolint: errcheck
func readUserInputFromEditor() []byte {
	tempfilename, err := launchUserEditor()
	if err != nil {
		log.Fatalf("couldn't open an editor: %v", err)
	}
	defer os.Remove(tempfilename)
	bs, err := ioutil.ReadFile(tempfilename)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	return bs
}

// nolint: errcheck
func launchUserEditor() (string, error) {
	editor, err := exec.LookPath("editor")
	if err != nil {
		editor = ""
	}
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		return "", errors.New("no editor set, make sure that either " +
			"VISUAL or EDITOR variables is set and pointing to a correct editor")
	}

	tempfile, err := ioutil.TempFile("", "sdkch_*")
	tempfilename := tempfile.Name()
	if err != nil {
		return "", err
	}
	tempfile.Close()

	cmd := exec.Command(editor, tempfilename)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		os.Remove(tempfilename)
		return "", err
	}

	fileInfo, err := os.Stat(tempfilename)
	if err != nil {
		os.Remove(tempfilename)
		return "", err
	}
	if fileInfo.Size() == 0 {
		return "", errors.New("aborting due to empty message")
	}

	return tempfilename, nil
}

func mustPruneDirIfEmpty(path string) {
	if contents := directoryContents(path); len(contents) == 0 {
		if err := os.Remove(path); err != nil {
			if !os.IsNotExist(err) {
				log.Fatal(err)
			}
			return
		}
		log.Println(path, "removed")
	}
}

// DONTCOVER
