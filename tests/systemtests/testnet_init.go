package systemtests

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cometbft/cometbft/p2p"
	"github.com/creachadair/tomledit"
	"github.com/creachadair/tomledit/parser"
)

// isV2 checks if the tests run with simapp v2
func isV2() bool {
	buildOptions := os.Getenv("COSMOS_BUILD_OPTIONS")
	return strings.Contains(buildOptions, "v2")
}

// SingleHostTestnetCmdInitializer default testnet cmd that supports the --single-host param
type SingleHostTestnetCmdInitializer struct {
	execBinary        string
	workDir           string
	chainID           string
	outputDir         string
	initialNodesCount int
	minGasPrice       string
	commitTimeout     time.Duration
	log               func(string)
}

// NewSingleHostTestnetCmdInitializer constructor
func NewSingleHostTestnetCmdInitializer(
	execBinary, workDir, chainID, outputDir string,
	initialNodesCount int,
	minGasPrice string,
	commitTimeout time.Duration,
	log func(string),
) *SingleHostTestnetCmdInitializer {
	return &SingleHostTestnetCmdInitializer{
		execBinary:        execBinary,
		workDir:           workDir,
		chainID:           chainID,
		outputDir:         outputDir,
		initialNodesCount: initialNodesCount,
		minGasPrice:       minGasPrice,
		commitTimeout:     commitTimeout,
		log:               log,
	}
}

// InitializerWithBinary creates new SingleHostTestnetCmdInitializer from sut with given binary
func InitializerWithBinary(binary string, sut *SystemUnderTest) TestnetInitializer {
	return NewSingleHostTestnetCmdInitializer(
		binary,
		WorkDir,
		sut.chainID,
		sut.outputDir,
		sut.initialNodesCount,
		sut.minGasPrice,
		sut.CommitTimeout(),
		sut.Log,
	)
}

func (s SingleHostTestnetCmdInitializer) Initialize() {
	args := []string{
		"testnet",
		"init-files",
		"--chain-id=" + s.chainID,
		"--output-dir=" + s.outputDir,
		"--validator-count=" + strconv.Itoa(s.initialNodesCount),
		"--keyring-backend=test",
		"--commit-timeout=" + s.commitTimeout.String(),
		"--single-host",
	}

	if isV2() {
		args = append(args, "--server.minimum-gas-prices="+s.minGasPrice)
	} else {
		args = append(args, "--minimum-gas-prices="+s.minGasPrice)
	}

	s.log(fmt.Sprintf("+++ %s %s\n", s.execBinary, strings.Join(args, " ")))
	out, err := RunShellCmd(s.execBinary, args...)
	if err != nil {
		panic(err)
	}
	s.log(out)
}

// ModifyConfigYamlInitializer testnet cmd prior to --single-host param. Modifies the toml files.
type ModifyConfigYamlInitializer struct {
	execBinary        string
	workDir           string
	chainID           string
	outputDir         string
	initialNodesCount int
	minGasPrice       string
	commitTimeout     time.Duration
	log               func(string)
	projectName       string
}

func NewModifyConfigYamlInitializer(exec string, s *SystemUnderTest) *ModifyConfigYamlInitializer {
	return &ModifyConfigYamlInitializer{
		execBinary:        exec,
		workDir:           WorkDir,
		chainID:           s.chainID,
		outputDir:         s.outputDir,
		initialNodesCount: s.initialNodesCount,
		minGasPrice:       s.minGasPrice,
		commitTimeout:     s.CommitTimeout(),
		log:               s.Log,
		projectName:       s.projectName,
	}
}

const (
	rpcPortStart  = 26657
	apiPortStart  = 1317
	grpcPortStart = 9090
	p2pPortStart  = 16656
)

func (s ModifyConfigYamlInitializer) Initialize() {
	// init with legacy testnet command
	args := []string{
		"testnet",
		"init-files",
		"--chain-id=" + s.chainID,
		"--output-dir=" + s.outputDir,
		"--v=" + strconv.Itoa(s.initialNodesCount),
		"--keyring-backend=test",
	}

	if isV2() {
		args = append(args, "--server.minimum-gas-prices="+s.minGasPrice)
	} else {
		args = append(args, "--minimum-gas-prices="+s.minGasPrice)
	}

	s.log(fmt.Sprintf("+++ %s %s\n", s.execBinary, strings.Join(args, " ")))

	out, err := RunShellCmd(s.execBinary, args...)
	if err != nil {
		panic(err)
	}
	s.log(out)

	nodeAddresses := make([]string, s.initialNodesCount)
	for i := 0; i < s.initialNodesCount; i++ {
		nodeDir := filepath.Join(WorkDir, NodePath(i, s.outputDir, s.projectName), "config")
		id := string(mustV(p2p.LoadNodeKey(filepath.Join(nodeDir, "node_key.json"))).ID())
		nodeAddresses[i] = fmt.Sprintf("%s@127.0.0.1:%d", id, p2pPortStart+i)
	}

	// then update configs
	for i := 0; i < s.initialNodesCount; i++ {
		nodeDir := filepath.Join(WorkDir, NodePath(i, s.outputDir, s.projectName), "config")
		nodeNumber := i
		EditToml(filepath.Join(nodeDir, "config.toml"), func(doc *tomledit.Document) {
			UpdatePort(doc, rpcPortStart+i, "rpc", "laddr")
			UpdatePort(doc, p2pPortStart+i, "p2p", "laddr")
			SetBool(doc, false, "p2p", "addr_book_strict")
			SetBool(doc, false, "p2p", "pex")
			SetBool(doc, true, "p2p", "allow_duplicate_ip")
			peers := make([]string, s.initialNodesCount)
			copy(peers, nodeAddresses[0:nodeNumber])
			copy(peers[nodeNumber:], nodeAddresses[nodeNumber+1:])
			SetValue(doc, strings.Join(peers, ","), "p2p", "persistent_peers")
			SetValue(doc, s.commitTimeout.String(), "consensus", "timeout_commit")
		})
		EditToml(filepath.Join(nodeDir, "app.toml"), func(doc *tomledit.Document) {
			UpdatePort(doc, apiPortStart+i, "api", "address")
			UpdatePort(doc, grpcPortStart+i, "grpc", "address")
		})
	}
}

func EditToml(filename string, f func(doc *tomledit.Document)) {
	tomlFile := mustV(os.OpenFile(filename, os.O_RDWR, 0o600))
	defer tomlFile.Close()
	doc := mustV(tomledit.Parse(tomlFile))
	f(doc)
	mustV(tomlFile.Seek(0, 0)) // reset the cursor to the beginning of the file
	must(tomlFile.Truncate(0))
	must(tomledit.Format(tomlFile, doc))
}

func SetBool(doc *tomledit.Document, newVal bool, xpath ...string) {
	e := doc.First(xpath...)
	if e == nil {
		panic(fmt.Sprintf("not found: %v", xpath))
	}
	e.Value = parser.MustValue(strconv.FormatBool(newVal))
}

func SetValue(doc *tomledit.Document, newVal string, xpath ...string) {
	e := doc.First(xpath...)
	if e == nil {
		panic(fmt.Sprintf("not found: %v", xpath))
	}
	e.Value = parser.MustValue(fmt.Sprintf("%q", newVal))
}

func UpdatePort(doc *tomledit.Document, newPort int, xpath ...string) {
	e := doc.First(xpath...)
	if e == nil {
		panic(fmt.Sprintf("not found: %v", xpath))
	}
	data := e.Value.X.String()
	pos := strings.LastIndexAny(data, ":")
	if pos == -1 {
		panic("column not found")
	}
	data = data[0:pos+1] + strconv.Itoa(newPort)
	e.Value = parser.MustValue(data + "\"")
}

// mustV same as must but with value returned
func mustV[T any](r T, err error) T {
	must(err)
	return r
}

// must simple panic on error for fluent calls
func must(err error) {
	if err != nil {
		panic(err)
	}
}
