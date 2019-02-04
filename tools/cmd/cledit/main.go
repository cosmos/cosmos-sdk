package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Section struct {
	GaiaREST   []string
	Gaiacli    []string
	Gaia       []string
	SDK        []string
	Tendermint []string
}

func NewSection() *Section {
	return &Section{GaiaREST: []string{}, Gaiacli: []string{}, Gaia: []string{}, SDK: []string{}}
}

func (se *Section) Empty() bool {
	return se == nil || (len(se.GaiaREST) == 0 &&
		len(se.Gaiacli) == 0 && len(se.Gaia) == 0 &&
		len(se.SDK) == 0)
}

func (se Section) GetStanza(name string) ([]string, error) {
	switch name {
	case "gaiarest":
		return se.GaiaREST, nil
	case "gaiacli":
		return se.Gaiacli, nil
	case "gaia":
		return se.Gaia, nil
	case "sdk":
		return se.SDK, nil
	case "tendermint":
		return se.Tendermint, nil
	}
	return nil, errors.New("unknown stanza")
}

func (se *Section) AppendItem(stanza, item string) error {
	switch stanza {
	case "gaiarest":
		se.GaiaREST = append(se.GaiaREST, item)
	case "gaiacli":
		se.Gaiacli = append(se.Gaiacli, item)
	case "gaia":
		se.Gaia = append(se.Gaia, item)
	case "sdk":
		se.SDK = append(se.SDK, item)
	case "tendermint":
		se.Tendermint = append(se.Tendermint, item)
	default:
		return errors.New("unknown stanza")
	}
	return nil
}

type Release struct {
	Breaking     *Section
	Features     *Section
	Improvements *Section
	Bugfixes     *Section
}

func NewRelease() *Release {
	return &Release{Breaking: NewSection(), Features: NewSection(), Improvements: NewSection(), Bugfixes: NewSection()}
}

func (r Release) GetSection(name string) (*Section, error) {
	switch name {
	case "breaking":
		return r.Breaking, nil
	case "improvements":
		return r.Improvements, nil
	case "features":
		return r.Features, nil
	case "bugfixes":
		return r.Bugfixes, nil
	}
	return nil, errors.New("unknown section")
}

var (
	progName string

	overwriteSourceFile bool
)

func init() {
	progName = filepath.Base(os.Args[0])
	flag.BoolVar(&overwriteSourceFile, "w", false, "write result to (source) file instead of stdout")
	flag.Usage = printUsage
}

func errInsufficientArgs() {
	log.Fatalf("insufficient arguments\nTry '%s -help' for more information.", progName)
}

func errTooManyArgs() {
	log.Fatalf("too many arguments\nTry '%s -help' for more information.", progName)
}

func unknownCommand(cmd string) {
	log.Fatalf("unknown command -- '%s'\nTry '%s -help' for more information.", cmd, progName)
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", filepath.Base(progName)))
	flag.Parse()

	if flag.NArg() < 1 {
		errInsufficientArgs()
	}

	cmd := flag.Arg(0)
	switch cmd {
	case "new":
		newFile()
		return
	case "add":
		switch {
		case flag.NArg() < 4:
			errInsufficientArgs()
		case flag.NArg() > 4:
			errTooManyArgs()
		}
		editFile(flag.Arg(1), flag.Arg(2), flag.Arg(3))
		return
	case "convert":
		switch {
		case flag.NArg() < 3:
			errInsufficientArgs()
		case flag.NArg() > 3:
			errTooManyArgs()
		}
		convert(flag.Arg(1), flag.Arg(2))
		return
	default:
		unknownCommand(cmd)
	}
}

func newFile() { fmt.Printf("%s", mustMarshal(NewRelease())) }

func editFile(clFile, section, stanza string) {
	r := unmarshalChangelogFile(clFile)

	releaseSection, err := r.GetSection(section)
	if err != nil {
		log.Fatalf("unknown section %q, possible values are %s", section,
			[]string{"breaking", "features", "improvements", "bugfixes"})
	}

	if _, err := releaseSection.GetStanza(stanza); err != nil {
		log.Fatalf("unknown stanza %q, possible values are %s", stanza,
			[]string{"gaia", "gaiacli", "gaiarest", "sdk", "tendermint"})
	}

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err := releaseSection.AppendItem(stanza, strings.TrimSpace(string(bytes))); err != nil {
		panic(err)
	}

	out, err := yaml.Marshal(r)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	outFile := os.Stdout
	if overwriteSourceFile {
		outFile, err = os.OpenFile(clFile, os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Fprintf(outFile, "%s", out)
}

func convert(clFile, version string) {
	r := unmarshalChangelogFile(clFile)
	md := fmt.Sprintf("## %s\n", version)
	md += processSection(r.Breaking, "BREAKING CHANGES")
	md += processSection(r.Features, "FEATURES")
	md += processSection(r.Improvements, "IMPROVEMENTS")
	md += processSection(r.Bugfixes, "BUGFIXES")
	fmt.Println(md)
}

func processSection(section *Section, header string) string {
	if section.Empty() {
		return ""
	}
	s := fmt.Sprintf("### %s\n", header)
	s += processStanza(section.GaiaREST, "Gaia REST API (`gaiacli rest-server`)")
	s += processStanza(section.Gaiacli, "Gaia CLI (`gaiacli`)")
	s += processStanza(section.Gaia, "Gaia")
	s += processStanza(section.SDK, "SDK")
	s += processStanza(section.Tendermint, "Tendermint")
	return s
}

func processStanza(stanza []string, header string) string {
	// regex to beautify github issues URLs
	if len(stanza) == 0 {
		return ""
	}
	s := fmt.Sprintf("* %s\n", header)
	for _, item := range stanza {
		s += processLine(item) + "\n"
	}
	return s
}

func processLine(s string) string {
	linesSlice := strings.Split(s, "\n")
	var out string

	if len(linesSlice) == 1 {
		return fmt.Sprintf("  * %s\n", expandGhURLs(linesSlice[0]))
	}
	for i, line := range linesSlice {
		line = strings.Trim(line, "\n")
		if i == 0 {
			out = fmt.Sprintf("  * %s\n", expandGhURLs(line))
		} else {
			out += fmt.Sprintf("    %s\n", expandGhURLs(line))
		}
	}
	return out
}

var reGhIssue = regexp.MustCompilePOSIX(`issue#([0-9]+)`)
var reGhPR = regexp.MustCompilePOSIX(`pr#([0-9]+)`)

func expandGhURLs(s string) string {
	return reGhPR.ReplaceAllString(reGhIssue.ReplaceAllString(s,
		"[\\#$1](https://github.com/cosmos/cosmos-sdk/issues/$1)"),
		"[\\#$1](https://github.com/cosmos/cosmos-sdk/pull/$1)")
}

func unmarshalChangelogFile(clFile string) *Release {
	contents, err := ioutil.ReadFile(clFile)
	if err != nil {
		log.Fatal(err)
	}

	r := NewRelease()
	if err := yaml.Unmarshal(contents, r); err != nil {
		log.Fatal(err)
	}
	return r
}

func mustMarshal(t interface{}) []byte {
	out, err := yaml.Marshal(t)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func printUsage() {
	usageText := fmt.Sprintf(`usage: %s [-w] [option]

Commands:
new                               Create new empty changelog.
add [-w] FILE SECTION STANZA      Add entry to a changelog file.
                                  Read from stdin until it
                                  encounters EOF.
convert FILE VERSION              Convert a changelog into
                                  Markdown format and print it
                                  to stdout.

    Sections:            Stanzas:
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
