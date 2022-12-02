package genesis

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"cosmossdk.io/core/internal/testpb"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/protobuf/proto"
)

func TestFileGenesisSourceOpenReaderWithField(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "test"
	field := "test"

	err := os.MkdirAll(filepath.Join(tmpdir, moduleName), dirCreateMode)
	require.NoError(t, err)

	fm := filepath.Join(tmpdir, moduleName, fmt.Sprintf("%s.json", field))
	f, err := os.Create(fm)
	require.NoError(t, err)
	defer f.Close()

	in := []byte("genesis read test!")
	_, err = f.Write(in)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, in, got)
}

func TestFileGenesisSourceOpenReaderWithCachedmoduleJson(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "test"
	field := "test"

	err := os.MkdirAll(filepath.Join(tmpdir, moduleName), dirCreateMode)
	require.NoError(t, err)

	fm := filepath.Join(tmpdir, moduleName, fmt.Sprintf("%s.json", field))
	f, err := os.Create(fm)
	require.NoError(t, err)
	defer f.Close()

	in := []byte("genesis read test!")
	_, err = f.Write(in)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	require.Equal(t, in, got)
}

func TestFileGenesisSourceOpenReaderWithModule(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"
	field := "field"

	fm := filepath.Join(tmpdir, fmt.Sprintf("%s.json", moduleName))
	f, err := os.Create(fm)
	require.NoError(t, err)
	defer f.Close()

	moduleState := json.RawMessage(`{"field":"value"}`)
	_, err = f.Write(moduleState)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)

	expected := json.RawMessage(`"value"`)
	require.Equal(t, expected, json.RawMessage(got))

	// verify the fileGenesisSource cached the moduleRawJSON
	fgs := gs.(*FileGenesisSource)
	require.Equal(t, moduleState, fgs.moduleRootJson)
}

func TestFileGenesisSourceOpenReaderWithGenesis(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"
	field := "field"

	fm := filepath.Join(tmpdir, "genesis.json")
	f, err := os.Create(fm)
	require.NoError(t, err)
	defer f.Close()

	appState := make(map[string]json.RawMessage)
	appState[moduleName] = json.RawMessage(`{"field":"value"}`)

	gc := tmtypes.GenesisDoc{}
	gc.AppState, err = json.Marshal(appState)
	require.NoError(t, err)

	gcRawJson, err := json.Marshal(gc)
	require.NoError(t, err)

	_, err = f.Write(gcRawJson)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)

	expected := json.RawMessage(`"value"`)
	require.Equal(t, expected, json.RawMessage(got))
}

func TestFileGenesisSourceOpenReaderWithInvalidFile(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "test"
	field := "test"

	fm := filepath.Join(tmpdir, "invalid_genesis.json")
	f, err := os.Create(fm)
	require.NoError(t, err)
	defer f.Close()

	in := []byte("genesis read test!")
	_, err = f.Write(in)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)
	_, err = gs.OpenReader(field)
	require.ErrorContains(t, err, "no such file or directory")
}

func TestFileGenesisSourceReadRawJSON(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	fp := filepath.Join(filepath.Clean(tmpdir), fmt.Sprintf("%s.json", moduleName))

	f, err := os.Create(fp)
	require.NoError(t, err)
	defer f.Close()

	in := []byte(`{"genesis": "read test"}`)
	_, err = f.Write(in)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)

	rj, err := gs.ReadRawJSON()
	require.NoError(t, err)
	require.Equal(t, in, []byte(rj))
}

func TestFileGenesisSourceReadRawJSONDefault(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	fp := filepath.Join(filepath.Clean(tmpdir), "genesis.json")

	f, err := os.Create(fp)
	require.NoError(t, err)
	defer f.Close()

	in := []byte(`{"genesis": "read test"}`)
	_, err = f.Write(in)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)

	rj, err := gs.ReadRawJSON()
	require.NoError(t, err)
	require.Equal(t, in, []byte(rj))
}

func TestFileGenesisSourceReadRawJSONWithEmptyModuleName(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := ""

	fp := filepath.Join(filepath.Clean(tmpdir), "genesis.json")

	f, err := os.Create(fp)
	require.NoError(t, err)
	defer f.Close()

	in := []byte(`{"genesis": "read test"}`)
	_, err = f.Write(in)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)

	_, err = gs.ReadRawJSON()
	require.ErrorContains(t, err, "failed to read RawJSON: empty module name")
}

func TestFileGenesisSourceReadMessage(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	fp := filepath.Join(filepath.Clean(tmpdir), "genesis.json")

	f, err := os.Create(fp)
	require.NoError(t, err)
	defer f.Close()

	m := &testpb.TestGenesisFile{}
	m.Key = "key"
	m.Value = "value"

	bz, err := proto.Marshal(m)
	require.NoError(t, err)

	_, err = f.Write(bz)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)

	mr := &testpb.TestGenesisFile{}

	err = gs.ReadMessage(mr)
	require.NoError(t, err)

	require.Equal(t, m.String(), mr.String())
}

func TestFileGenesisTargetWithField(t *testing.T) {
	tmp := t.TempDir()

	moduleName := ""
	field := "field"

	// empty module name, should return error
	gs := NewFileGenesisTarget(tmp, moduleName)
	_, err := gs.OpenWriter(field)
	require.Error(t, err)

	// open file with module field
	moduleName = "module"
	gs = NewFileGenesisTarget(tmp, moduleName)
	w, err := gs.OpenWriter(field)
	require.NoError(t, err)
	defer w.Close()

	msg := []byte("test message")
	_, err = w.Write(msg)
	require.NoError(t, err)

	fp, err := os.Open(filepath.Join(tmp, moduleName, fmt.Sprintf("%s.json", field)))
	require.NoError(t, err)
	defer fp.Close()

	got, err := io.ReadAll(fp)
	require.NoError(t, err)
	require.Equal(t, msg, got)
}

func TestFileGenesisTargetWithModule(t *testing.T) {
	tmp := t.TempDir()

	moduleName := "test"
	field := ""

	gs := NewFileGenesisTarget(tmp, moduleName)
	w, err := gs.OpenWriter(field)
	require.NoError(t, err)
	defer w.Close()

	msg := []byte("test message")
	_, err = w.Write(msg)
	require.NoError(t, err)

	fp, err := os.Open(filepath.Join(tmp, fmt.Sprintf("%s.json", moduleName)))
	require.NoError(t, err)
	defer fp.Close()

	got, err := io.ReadAll(fp)
	require.NoError(t, err)
	require.Equal(t, msg, got)
}

func TestFileGenesisTargetWithoutModuleAndField(t *testing.T) {
	tmp := t.TempDir()

	moduleName := ""
	field := ""

	gs := NewFileGenesisTarget(tmp, moduleName)
	w, err := gs.OpenWriter(field)
	require.NoError(t, err)
	defer w.Close()

	msg := []byte("test message")
	_, err = w.Write(msg)
	require.NoError(t, err)

	fp, err := os.Open(filepath.Join(tmp, "genesis.json"))
	require.NoError(t, err)
	defer fp.Close()

	got, err := io.ReadAll(fp)
	require.NoError(t, err)
	require.Equal(t, msg, got)
}

func TestFileGenesisTargetWriteMessageNoOp(t *testing.T) {
	var pm proto.Message
	gs := NewFileGenesisTarget("", "")
	err := gs.WriteMessage(pm)
	require.Error(t, err, "unsupported op")
}

func TestFileGenesisTargetWriteRawJSONEmptyModuleName(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := ""

	gt := NewFileGenesisTarget(tmpdir, moduleName)

	str := []byte(`{"genesis": "write test"}`)
	err := gt.WriteRawJSON(str)
	require.ErrorContains(t, err, "failed to write RawJSON: empty module name")
}

func TestFileGenesisTargetWriteRawJSON(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	gt := NewFileGenesisTarget(tmpdir, moduleName)

	in := []byte(`{"genesis": "write test"}`)
	err := gt.WriteRawJSON(in)
	require.NoError(t, err)

	fp := filepath.Join(filepath.Clean(tmpdir), fmt.Sprintf("%s.json", moduleName))

	f, err := os.Open(fp)
	require.NoError(t, err)
	defer f.Close()

	got, err := io.ReadAll(f)
	require.NoError(t, err)
	require.Equal(t, in, got)
}

func TestFileGenesisTargetWriteRawJSONWithIndent(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	gt := NewFileGenesisTargetWithIndent(tmpdir, moduleName)

	in := []byte(`{"genesis": "write test"}`)
	err := gt.WriteRawJSON(in)
	require.NoError(t, err)

	fp := filepath.Join(filepath.Clean(tmpdir), fmt.Sprintf("%s.json", moduleName))

	f, err := os.Open(fp)
	require.NoError(t, err)
	defer f.Close()

	got, err := io.ReadAll(f)
	require.NoError(t, err)

	fstr, err := json.MarshalIndent(json.RawMessage(in), "", "  ")
	require.NoError(t, err)
	require.Equal(t, fstr, got)
}
