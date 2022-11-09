package genesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
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

	str := "genesis read test!"
	_, err = f.WriteString(str)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	require.NoError(t, err)

	require.Equal(t, str, buf.String())
}

func TestFileGenesisSourceOpenReaderWithModule(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "test"
	field := "test"

	fm := filepath.Join(tmpdir, fmt.Sprintf("%s.json", moduleName))
	f, err := os.Create(fm)
	require.NoError(t, err)
	defer f.Close()

	str := "genesis read test!"
	_, err = f.WriteString(str)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	require.NoError(t, err)

	require.Equal(t, str, buf.String())
}

func TestFileGenesisSourceOpenReaderWithGenesis(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "test"
	field := "test"

	fm := filepath.Join(tmpdir, "genesis.json")
	f, err := os.Create(fm)
	require.NoError(t, err)
	defer f.Close()

	str := "genesis read test!"
	_, err = f.WriteString(str)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)
	reader, err := gs.OpenReader(field)
	require.NoError(t, err)
	defer reader.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(reader)
	require.NoError(t, err)

	require.Equal(t, str, buf.String())
}

func TestFileGenesisSourceOpenReaderWithInvalidFile(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "test"
	field := "test"

	fm := filepath.Join(tmpdir, "invalid_genesis.json")
	f, err := os.Create(fm)
	require.NoError(t, err)
	defer f.Close()

	str := "genesis read test!"
	_, err = f.WriteString(str)
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

	str := `{"genesis": "read test"}`

	_, err = f.WriteString(str)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)

	rj, err := gs.ReadRawJSON()
	require.NoError(t, err)

	require.Equal(t, str, (string)(rj))
}

func TestFileGenesisSourceReadRawJSONDefault(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	fp := filepath.Join(filepath.Clean(tmpdir), "genesis.json")

	f, err := os.Create(fp)
	require.NoError(t, err)
	defer f.Close()

	str := `{"genesis": "read test"}`

	_, err = f.WriteString(str)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)

	rj, err := gs.ReadRawJSON()
	require.NoError(t, err)

	require.Equal(t, str, (string)(rj))
}

func TestFileGenesisSourceReadRawJSONWithEmptyModuleName(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := ""

	fp := filepath.Join(filepath.Clean(tmpdir), "genesis.json")

	f, err := os.Create(fp)
	require.NoError(t, err)
	defer f.Close()

	str := `{"genesis": "read test"}`

	_, err = f.WriteString(str)
	require.NoError(t, err)

	gs := NewFileGenesisSource(tmpdir, moduleName)

	_, err = gs.ReadRawJSON()
	require.ErrorContains(t, err, "failed to write RawJSON: empty module name")
}

func TestFileGenesisSourceReadMessageNoOp(t *testing.T) {
	gs := NewFileGenesisSource("", "")

	var pm proto.Message
	err := gs.ReadMessage(pm)
	require.Error(t, err, "unsupportted op")
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

	msg := "test message"
	_, err = w.Write([]byte(msg))
	require.NoError(t, err)

	fp, err := os.Open(filepath.Join(tmp, moduleName, fmt.Sprintf("%s.json", field)))
	require.NoError(t, err)
	defer fp.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(fp)
	require.NoError(t, err)

	require.Equal(t, msg, buf.String())
}

func TestFileGenesisTargetWithModule(t *testing.T) {
	tmp := t.TempDir()

	moduleName := "test"
	field := ""

	gs := NewFileGenesisTarget(tmp, moduleName)
	w, err := gs.OpenWriter(field)
	require.NoError(t, err)
	defer w.Close()

	msg := "test message"
	_, err = w.Write([]byte(msg))
	require.NoError(t, err)

	fp, err := os.Open(filepath.Join(tmp, fmt.Sprintf("%s.json", moduleName)))
	require.NoError(t, err)
	defer fp.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(fp)
	require.NoError(t, err)

	require.Equal(t, msg, buf.String())
}

func TestFileGenesisTargetWithoutModuleAndField(t *testing.T) {
	tmp := t.TempDir()

	moduleName := ""
	field := ""

	gs := NewFileGenesisTarget(tmp, moduleName)
	w, err := gs.OpenWriter(field)
	require.NoError(t, err)
	defer w.Close()

	msg := "test message"
	_, err = w.Write([]byte(msg))
	require.NoError(t, err)

	fp, err := os.Open(filepath.Join(tmp, "genesis.json"))
	require.NoError(t, err)
	defer fp.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(fp)
	require.NoError(t, err)

	require.Equal(t, msg, buf.String())
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

	str := `{"genesis": "write test"}`
	err := gt.WriteRawJSON([]byte(str))
	require.ErrorContains(t, err, "failed to write RawJSON: empty module name")
}

func TestFileGenesisTargetWriteRawJSON(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	gt := NewFileGenesisTarget(tmpdir, moduleName)

	str := `{"genesis": "write test"}`
	err := gt.WriteRawJSON([]byte(str))
	require.NoError(t, err)

	fp := filepath.Join(filepath.Clean(tmpdir), fmt.Sprintf("%s.json", moduleName))

	f, err := os.Open(fp)
	require.NoError(t, err)
	defer f.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(f)
	require.NoError(t, err)

	require.Equal(t, str, buf.String())
}

func TestFileGenesisTargetWriteRawJSONWithIndent(t *testing.T) {
	tmpdir := t.TempDir()
	moduleName := "module"

	gt := NewFileGenesisTargetWithIndent(tmpdir, moduleName)

	str := `{"genesis": "write test"}`
	err := gt.WriteRawJSON([]byte(str))
	require.NoError(t, err)

	fp := filepath.Join(filepath.Clean(tmpdir), fmt.Sprintf("%s.json", moduleName))

	f, err := os.Open(fp)
	require.NoError(t, err)
	defer f.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(f)
	require.NoError(t, err)

	fstr, err := json.MarshalIndent(json.RawMessage(str), "", "  ")
	require.NoError(t, err)

	require.Equal(t, string(fstr), buf.String())
}
