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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var testModuleState = json.RawMessage(`{"module":"state"}`)

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

	gs := NewFileGenesisSource(tmpdir, moduleName, testModuleState)
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

	gs := NewFileGenesisSource(tmpdir, moduleName, testModuleState)
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)

	expected := json.RawMessage(`"value"`)
	require.Equal(t, expected, json.RawMessage(got))
}

func TestFileGenesisSourceOpenReaderWithGenesisAppState(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"
	field := "field"

	gs := NewFileGenesisSource(tmpdir, moduleName, json.RawMessage(`{"field":"value"}`))
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)

	expected := json.RawMessage(`"value"`)
	require.Equal(t, expected, json.RawMessage(got))
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

	gs := NewFileGenesisSource(tmpdir, moduleName, testModuleState)

	rj, err := gs.ReadRawJSON()
	require.NoError(t, err)
	require.Equal(t, in, []byte(rj))
}

func TestFileGenesisSourceReadRawJSONNoFileExist(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	gs := NewFileGenesisSource(tmpdir, moduleName, testModuleState)

	bz, err := gs.ReadRawJSON()
	require.NoError(t, err)
	require.Equal(t, testModuleState, bz)
}

func TestFileGenesisSourceReadMessage(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	in := &testpb.TestGenesisFile{}
	in.Key = "key"
	in.Value = "value"

	bz, err := protojson.Marshal(in)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName, bz)

	out := &testpb.TestGenesisFile{}
	err = gs.ReadMessage(out)
	require.NoError(t, err)

	require.Equal(t, in.String(), out.String())
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
