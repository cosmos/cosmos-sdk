package main

import (
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

type Stanza []string

func NewStanza() *Stanza { return new(Stanza) }

func (st *Stanza) Empty() bool { return st == nil || len(*st) == 0 }

func (st *Stanza) Append(s string) { *st = append(*st, s) }

type Section struct {
	GaiaREST   *Stanza
	Gaiacli    *Stanza
	Gaia       *Stanza
	SDK        *Stanza
	Tendermint *Stanza
}

func NewSection() *Section {
	return &Section{GaiaREST: NewStanza(), Gaiacli: NewStanza(), Gaia: NewStanza(), SDK: NewStanza()}
}

func (se *Section) Empty() bool {
	return se == nil || (se.GaiaREST.Empty() && se.Gaiacli.Empty() && se.Gaia.Empty() && se.SDK.Empty())
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

var progName string

func init() {
	progName = filepath.Base(os.Args[0])
	flag.Bool("w", false, "write result to (source) file instead of stdout")
	flag.Usage = func() {
		usageText := fmt.Sprintf(`usage: %s [-w] [option]

Commands:
    new                          Create new empty changelog.
    add FILE SECTION STANZA      Add entry to a changelog file.
                                 Read from stdin until it
							     encounters EOF.
	convert FILE VERSION         Convert a changelog into
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
}

func main() {
	log.SetFlags(0)
	log.SetPrefix(fmt.Sprintf("%s: ", filepath.Base(progName)))
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("insufficient arguments")
	}

	cmd := flag.Arg(0)
	switch cmd {
	case "new":
		newFile()
		return
	case "add":
		switch {
		case flag.NArg() < 4:
			log.Fatal("insufficient arguments")
		case flag.NArg() > 4:
			log.Fatal("too many arguments")
		}
		editFile(flag.Arg(1), flag.Arg(2), flag.Arg(3))
		return
	case "convert":
		switch {
		case flag.NArg() < 3:
			log.Fatal("insufficient arguments")
		case flag.NArg() > 3:
			log.Fatal("too many arguments")
		}
		convert(flag.Arg(1), flag.Arg(2))
		return
	default:
		log.Fatal(fmt.Errorf("unknown command -- '%s'", cmd))
	}
}

func newFile() { fmt.Printf("%s", mustMarshal(Release{})) }

func editFile(clFile, section, stanza string) {
	r := unmarshalChangelogFile(clFile)

	var releaseSection *Section
	switch section {
	case "breaking":
		releaseSection = r.Breaking
	case "features":
		releaseSection = r.Features
	case "improvements":
		releaseSection = r.Improvements
	case "bugfixes":
		releaseSection = r.Bugfixes
	default:
		log.Fatalf("unknown section %q, possible values are %s", section,
			[]string{"breaking", "features", "improvements", "bugfixes"})
	}

	//var releaseStanza Stanza

	//fmt.Printf("%+v\n", releaseSection)

	releaseStanza := NewStanza()
	switch stanza {
	case "gaia":
		releaseStanza = releaseSection.Gaia
	case "gaiacli":
		releaseStanza = releaseSection.Gaiacli
	case "gaiarest":
		releaseStanza = releaseSection.GaiaREST
	case "sdk":
		releaseStanza = releaseSection.SDK
	case "tendermint":
		releaseStanza = releaseSection.Tendermint
	default:
		log.Fatalf("unknown stanza %q, possible values are %s", stanza,
			[]string{"gaia", "gaiacli", "gaiarest", "sdk", "tendermint"})
	}

	bytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	releaseStanza.Append(strings.TrimSpace(string(bytes)))

	out, err := yaml.Marshal(r)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("%s", out)
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

func processStanza(stanza *Stanza, header string) string {
	// regex to beautify github issues URLs
	if stanza.Empty() {
		return ""
	}
	s := fmt.Sprintf("* %s\n", header)
	for _, item := range *stanza {
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
