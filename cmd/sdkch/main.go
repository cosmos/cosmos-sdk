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

	entriesDir     string
	verboseLogging bool

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
)

func init() {
	progName = filepath.Base(os.Args[0])
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&entriesDir, "d", filepath.Join(cwd, entriesDirName), "entry files directory")
	flag.BoolVar(&verboseLogging, "v", false, "enable verbose logging")
	flag.Usage = printUsage

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

	if flag.NArg() < 1 {
		errInsufficientArgs()
	}

	cmd := flag.Arg(0)
	switch cmd {

	case "add":
		if flag.NArg() < 3 {
			errInsufficientArgs()
		}
		if flag.NArg() > 4 {
			errTooManyArgs()
		}
		sectionDir, stanzaDir := flag.Arg(1), flag.Arg(2)
		validateSectionStanzaDirs(sectionDir, stanzaDir)
		if flag.NArg() == 4 {
			addSinglelineEntryFile(sectionDir, stanzaDir, strings.TrimSpace(flag.Arg(3)))
			return
		}
		addEntryFile(sectionDir, stanzaDir)

	case "generate":
		version := "UNRELEASED"
		if flag.NArg() > 1 {
			version = strings.Join(flag.Args()[1:], " ")
		}
		generateChangelog(version)

	case "prune":
		pruneEmptyDirectories()

	default:
		unknownCommand(cmd)
	}
}

func addSinglelineEntryFile(sectionDir, stanzaDir, message string) {
	filename := filepath.Join(
		filepath.Join(entriesDir, sectionDir, stanzaDir),
		generateFileName(message),
	)

	writeEntryFile(filename, []byte(message))
}

func addEntryFile(sectionDir, stanzaDir string) {
	bs := readUserInput()
	firstLine := strings.TrimSpace(strings.Split(string(bs), "\n")[0])
	filename := filepath.Join(
		filepath.Join(entriesDir, sectionDir, stanzaDir),
		generateFileName(firstLine),
	)

	writeEntryFile(filename, bs)
}

var filenameInvalidChars = regexp.MustCompile(`[^a-zA-Z0-9-_]`)

func generateFileName(line string) string {
	var chunks []string
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
	for sectionDir, _ := range sections {
		for stanzaDir, _ := range stanzas {
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
func readUserInput() []byte {
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
		log.Fatal("aborting due to empty message")
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

func printUsage() {
	usageText := fmt.Sprintf(`usage: %s [-d directory] [-v] command

Maintain unreleased changelog entries in a modular fashion.

Commands:
    add [section] [stanza] [message]  Add an entry file. If message is empty, start 
                                      the editor to edit the message.
    generate [version]                Generate a changelog in Markdown format and print
                                      it to STDOUT. version defaults to UNRELEASED.
    prune                             Delete empty sub-directories recursively.

    Sections             Stanzas
         ---                 ---
    breaking                gaia
    features             gaiacli
improvements            gaiarest
    bugfixes                 sdk
                      tendermint
`, progName)
	fmt.Fprintf(os.Stderr, "%s\nFlags:\n", usageText)
	flag.PrintDefaults()
}

func errInsufficientArgs() {
	log.Println("insufficient arguments")
	printUsage()
	os.Exit(1)
}

func errTooManyArgs() {
	log.Println("too many arguments")
	printUsage()
	os.Exit(1)
}

func unknownCommand(cmd string) {
	log.Fatalf("unknown command -- '%s'\nTry '%s -help' for more information.", cmd, progName)
}

// DONTCOVER
