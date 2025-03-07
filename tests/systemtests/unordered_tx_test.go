//go:build system_test

package systemtests

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	systest "cosmossdk.io/systemtests"
)

func TestUnorderedTXDuplicate(t *testing.T) {
	// scenario: test unordered tx duplicate
	// given a running chain with a tx in the unordered tx pool
	// when a new tx with the same hash is broadcasted
	// then the new tx should be rejected

	systest.Sut.ResetChain(t)
	cli := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// add genesis account with some tokens
	account1Addr := cli.AddKey("account1")
	account2Addr := cli.AddKey("account2")
	systest.Sut.ModifyGenesisCLI(t,
		[]string{"genesis", "add-genesis-account", account1Addr, "10000000stake"},
	)

	systest.Sut.StartChain(t)

	timeoutTimestamp := time.Now().Add(time.Minute)
	// send tokens
	cmd := []string{"tx", "bank", "send", account1Addr, account2Addr, "5000stake", "--from=" + account1Addr, "--fees=1stake", fmt.Sprintf("--timeout-timestamp=%v", timeoutTimestamp.Unix()), "--unordered", "--sequence=1", "--note=1"}
	rsp1 := cli.Run(cmd...)
	systest.RequireTxSuccess(t, rsp1)

	assertDuplicateErr := func(xt assert.TestingT, gotErr error, gotOutputs ...interface{}) bool {
		require.Len(t, gotOutputs, 1)
		output := gotOutputs[0].(string)
		code := gjson.Get(output, "code")
		require.True(t, code.Exists())
		require.Equal(t, int64(19), code.Int()) // 19 == already in mempool.
		return false                            // always abort
	}
	rsp2 := cli.WithRunErrorMatcher(assertDuplicateErr).Run(cmd...)
	systest.RequireTxFailure(t, rsp2)

	require.Eventually(t, func() bool {
		return cli.QueryBalance(account2Addr, "stake") == 5000
	}, 10*systest.Sut.BlockTime(), 200*time.Millisecond, "TX was not executed before timeout")
}

func TestTxBackwardsCompatability(t *testing.T) {
	// Scenario:
	// A transaction generated from a v0.53 chain without unordered and timeout_timestamp flags set should succeed.
	// Conversely, a transaction generated from a v0.53 chain with unordered and timeout_timestamp flags set should fail.
	err := DownloadSimdBinary(t, "v0.50.5", systest.WorkDir+"/binaries/simdv50")
	require.NoError(t, err)

	var (
		denom                = "stake"
		transferAmount int64 = 1000
		testSeed             = "scene learn remember glide apple expand quality spawn property shoe lamp carry upset blossom draft reject aim file trash miss script joy only measure"
	)
	systest.Sut.ResetChain(t)

	v53CLI := systest.NewCLIWrapper(t, systest.Sut, systest.Verbose)
	// we just get val addr for an address to send things to.
	valAddr := v53CLI.GetKeyAddr("node0")
	require.NotEmpty(t, valAddr)

	// generate a deterministic account. we'll use this seed again later in the v50 chain.
	senderAddr := v53CLI.AddKeyFromSeed("account1", testSeed)

	//// Now we're going to switch to a v.50 chain.
	legacyBinary := systest.WorkDir + "/binaries/simdv50"

	// setup the v50 chain. v53 made some changes to testnet command, so we'll have to adjust here. and use only 1 node.
	legacySut := systest.NewSystemUnderTest(legacyBinary, systest.Verbose, 1, 1*time.Second)
	legacySut.SetTestnetInitializer(systest.LegacyInitializerWithBinary(legacyBinary, legacySut))
	legacySut.SetupChain()
	legacySut.SetExecBinary(legacyBinary) // doing this because of a bug. for some reason it sets my exec binary to the legacyBinary appended to itself. so (legacy/bin/path/legacy/bin/path)
	v50CLI := systest.NewCLIWrapper(t, legacySut, systest.Verbose)
	v50CLI.AddKeyFromSeed("account1", testSeed)
	legacySut.ModifyGenesisCLI(t,
		// add some bogus accounts because the v53 chain had 4 nodes which takes account numbers 1-4.
		[]string{"genesis", "add-genesis-account", v50CLI.AddKey("foo"), "10000000000stake"},
		[]string{"genesis", "add-genesis-account", v50CLI.AddKey("bar"), "10000000000stake"},
		[]string{"genesis", "add-genesis-account", v50CLI.AddKey("baz"), "10000000000stake"},
		// we need our sender to be account 5 because that's how it was signed in the v53 scenario.
		[]string{"genesis", "add-genesis-account", senderAddr, "10000000000stake"},
	)

	legacySut.StartChain(t)

	bankSendCmdArgs := []string{"tx", "bank", "send", senderAddr, valAddr, fmt.Sprintf("%d%s", transferAmount, denom), "--chain-id=" + v50CLI.ChainID(), "--fees=10stake", "--sign-mode=direct"}
	res, ok := v53CLI.RunOnly(bankSendCmdArgs...)
	require.True(t, ok)

	response, ok := v50CLI.AwaitTxCommitted(res, 15*time.Second)
	require.True(t, ok)
	code := gjson.Get(response, "code").Int()
	require.Equal(t, int64(0), code)
}

const (
	repoOwner = "cosmos"
	repoName  = "cosmos-sdk"
)

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

func DownloadSimdBinary(t *testing.T, version string, installPath string) error {
	tag := url.PathEscape("simd/" + version)
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", repoOwner, repoName, tag)

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to decode release response: %w", err)
	}

	osName, arch := runtime.GOOS, runtime.GOARCH
	targetName := fmt.Sprintf("simd-%s-%s-%s", version, osName, arch)

	var matchingAsset *Asset
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, targetName) {
			matchingAsset = &asset
			break
		}
	}

	if matchingAsset == nil {
		return fmt.Errorf("no matching binary found for %s/%s", osName, arch)
	}

	t.Logf("Downloading %s...", matchingAsset.Name)
	tmpArchive, err := os.CreateTemp("", matchingAsset.Name)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpArchive.Name())
	defer tmpArchive.Close()

	assetResp, err := http.Get(matchingAsset.BrowserDownloadURL)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer assetResp.Body.Close()

	if _, err = io.Copy(tmpArchive, assetResp.Body); err != nil {
		return fmt.Errorf("failed to save asset: %w", err)
	}

	extractDir, err := os.MkdirTemp("", "simd_extract_*")
	if err != nil {
		return fmt.Errorf("failed to create temp extract dir: %w", err)
	}
	defer os.RemoveAll(extractDir)

	t.Logf("Extracting to %s...", extractDir)
	if strings.HasSuffix(matchingAsset.Name, ".tar.gz") {
		err = extractTarGz(tmpArchive.Name(), extractDir)
	} else if strings.HasSuffix(matchingAsset.Name, ".zip") {
		err = extractZip(tmpArchive.Name(), extractDir)
	} else {
		return fmt.Errorf("unknown archive format: %s", matchingAsset.Name)
	}
	if err != nil {
		return fmt.Errorf("failed to extract archive: %w", err)
	}

	var simdBinaryPath string
	err = filepath.Walk(extractDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Mode().IsRegular() && info.Name() == "simd" {
			simdBinaryPath = path
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return fmt.Errorf("error finding simd binary: %w", err)
	}
	if simdBinaryPath == "" {
		return fmt.Errorf("simd binary not found in extracted files")
	}

	t.Logf("Installing simd binary to %s...", installPath)
	input, err := os.Open(simdBinaryPath)
	if err != nil {
		return fmt.Errorf("failed to open simd binary: %w", err)
	}
	defer input.Close()

	output, err := os.OpenFile(installPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("failed to create install file: %w", err)
	}
	defer output.Close()

	if _, err := io.Copy(output, input); err != nil {
		return fmt.Errorf("failed to install simd binary: %w", err)
	}

	t.Log("simd binary installed successfully!")
	return nil
}

func extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		filePath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(filePath, 0755); err != nil {
				return err
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
