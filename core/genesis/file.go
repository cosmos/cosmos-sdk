package genesis

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"google.golang.org/protobuf/proto"

	tmtypes "github.com/tendermint/tendermint/types"
)

type FileGenesisSource struct {
	sourceDir           string
	moduleName          string
	moduleRootJson      json.RawMessage // the RawMessage from the genesis.json app_state.<module> that got passed into InitCHain
	readFromRootGenesis bool
}

const (
	fileOpenflag  = os.O_CREATE | os.O_WRONLY
	flieOpenMode  = fs.FileMode(0o600)
	dirCreateMode = fs.FileMode(0o700)
)

// NewFileGenesisSource returns a new GenesisSource for the provided
// source directory and the provided module name where it is assumed
// that it contains encoded JSON data in the file.
func NewFileGenesisSource(sourceDir, moduleName string) GenesisSource {
	return &FileGenesisSource{sourceDir: filepath.Clean(sourceDir), moduleName: moduleName}
}

// OpenReader opens the source field reading from the given parameters,
// and returns a ReadCloser.
// It will try to open the field in order following by:
// <sourceDir>/<module>/<field>.json
// <field> key inside <sourceDir>/<module>.json
// app_state.<module>.<field> key in <sourceDir>/genesis.json
func (f *FileGenesisSource) OpenReader(field string) (io.ReadCloser, error) {
	var rawBz json.RawMessage

	// if moduleRootJson is not nil, we can skip reading data from file
	if f.moduleRootJson != nil {
		rawBz = f.moduleRootJson
	}

	if rawBz == nil {
		// try reading genesis data from <sourceDir>/<module>/<field>.json
		fName := fmt.Sprintf("%s.json", field)
		fPath := filepath.Join(f.sourceDir, f.moduleName)

		fp, err := os.Open(filepath.Clean(filepath.Join(fPath, fName)))
		if err == nil {
			return fp, nil
		}

		// if cannot find it, try reading from <sourceDir>/<module>.json
		// or <sourceDir>/genesis.json
		rawBz, err = f.ReadRawJSON()
		if err != nil {
			return nil, err
		}
	}

	if !f.readFromRootGenesis {
		// rawBz has been loaded from the <module>.json
		f.moduleRootJson = rawBz
		return f.unmarshalRawModuleWithField(rawBz, field)
	}

	// unmarshal module rawJSON from genesis.AppState
	doc := tmtypes.GenesisDoc{}
	err := json.Unmarshal(rawBz, &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rawJSON to GenesisDoc: %w", err)
	}

	appState := make(map[string]json.RawMessage)
	err = json.Unmarshal(doc.AppState, &appState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal app state from GenesisDoc: %w", err)
	}

	moduleState := appState[f.moduleName]
	if moduleState == nil {
		return nil, fmt.Errorf("failed to retrieve module state %s from genesis.json", f.moduleName)
	}
	f.moduleRootJson = moduleState

	return f.unmarshalRawModuleWithField(moduleState, field)
}

func (f *FileGenesisSource) unmarshalRawModuleWithField(rawBz []byte, field string) (io.ReadCloser, error) {
	fieldState := make(map[string]json.RawMessage)
	err := json.Unmarshal(rawBz, &fieldState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields from module state %s", f.moduleName)
	}

	fieldRawData := fieldState[field]
	if fieldRawData == nil {
		return nil, fmt.Errorf("failed to retrieve module field %s/%s from genesis.json", f.moduleName, field)
	}

	// wrap raw field data to a ReadCloser
	return io.NopCloser(bytes.NewReader(fieldRawData)), nil
}

// ReadMessage reads rawJSON data from source file and unmarshal it into proto.Message
func (f *FileGenesisSource) ReadMessage(msg proto.Message) error {
	bz, err := f.ReadRawJSON()
	if err != nil {
		return err
	}
	return proto.Unmarshal(bz, msg)
}

// ReadRawJSON returns a json.RawMessage read from the source file given by the
// source directory and the module name. If it cannot open <sourceDir>/<module>.json,
// it will try to read raw JSON data from <sourceDir>/genesis.json
func (f *FileGenesisSource) ReadRawJSON() (rawBz json.RawMessage, rerr error) {
	if len(f.moduleName) == 0 {
		return nil, fmt.Errorf("failed to read RawJSON: empty module name")
	}

	fName := fmt.Sprintf("%s.json", f.moduleName)
	fPath := filepath.Join(f.sourceDir, fName)

	fp, err := os.Open(filepath.Clean(fPath))
	if err != nil {
		// try reading from <sourceDir>/genesis.json if it's not able to read from
		// <sourceDir>/<moduleName>.json
		fPath = filepath.Join(f.sourceDir, "genesis.json")
		fp, err = os.Open(filepath.Clean(fPath))
		if err != nil {
			return nil, fmt.Errorf("failed to open file from %s: %w", fPath, err)
		}
		f.readFromRootGenesis = true
	}

	defer func() {
		if err := fp.Close(); err != nil {
			if rerr != nil {
				rerr = fmt.Errorf("failed to close file %s: %s, %w", fp.Name(), err.Error(), rerr)
				return
			}
			rerr = fmt.Errorf("failed to close file %s: %w", fp.Name(), err)
		}
	}()

	fi, err := fp.Stat()
	if err != nil {
		rerr = fmt.Errorf("failed to stat file %s: %w", fp.Name(), rerr)
		return nil, rerr
	}

	var buf bytes.Buffer
	n, err := buf.ReadFrom(fp)
	if err != nil {
		rerr = fmt.Errorf("failed to read file %s: %w", fp.Name(), err)
		return nil, rerr
	}

	if n != fi.Size() {
		rerr = fmt.Errorf("couldn't read entire file: %s, read: %d, file size: %d", fp.Name(), n, fi.Size())
		return nil, rerr
	}

	return buf.Bytes(), nil
}

type FileGenesisTarget struct {
	targetDir  string
	moduleName string
	indent     bool
}

// NewFileGenesisTarget returns GenesisTarget implementation with given target directory
// and the given module name.
func NewFileGenesisTarget(targetDir, moduleName string) GenesisTarget {
	return &FileGenesisTarget{
		targetDir:  filepath.Clean(targetDir),
		moduleName: moduleName,
	}
}

// NewFileGenesisTargetWithIndent returns GenesisTarget implementation with given target directory,
// the given module name, and enabled the indent option for JSON raw data.
func NewFileGenesisTargetWithIndent(targetDir, moduleName string) GenesisTarget {
	return &FileGenesisTarget{
		targetDir:  filepath.Clean(targetDir),
		moduleName: moduleName,
		indent:     true,
	}
}

// OpenWriter create a file for writing the genesus state to the file.
// It will try to create a file in order following by:
// <targetDir>/<module>/<field>.json
// <targetDir>/<module>.json when field is empty
// <targetDir>/genesis.json when both module name and field are empty
func (f *FileGenesisTarget) OpenWriter(field string) (io.WriteCloser, error) {
	// try to create open/create a file to <targetDir>/<module>/<field>.json
	if len(field) > 0 {
		if len(f.moduleName) == 0 {
			return nil, fmt.Errorf("failed to open writer, the module name must be specified when field is assigned")
		}

		fPath := filepath.Join(f.targetDir, f.moduleName)
		if err := os.MkdirAll(fPath, dirCreateMode); err != nil {
			return nil, fmt.Errorf("failed to create target directory %s: %w", fPath, err)
		}

		fileName := fmt.Sprintf("%s.json", field)
		fp, err := os.OpenFile(filepath.Clean(filepath.Join(f.targetDir, f.moduleName, fileName)), fileOpenflag, flieOpenMode)
		if err != nil {
			return nil, fmt.Errorf("failed to open writer, %s: %w", fileName, err)
		}
		return fp, nil
	}

	if err := os.MkdirAll(f.targetDir, dirCreateMode); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// if there is empty field, try to open/create a file to <targetDir>/<module>.json
	if len(f.moduleName) > 0 {
		fName := fmt.Sprintf("%s.json", f.moduleName)
		fp, err := os.OpenFile(filepath.Clean(filepath.Join(f.targetDir, fName)), fileOpenflag, flieOpenMode)
		if err != nil {
			return nil, fmt.Errorf("failed to open writer, %s: %v", fName, err)
		}
		return fp, nil
	}

	// else if there is empty module and field name try to open/create a file to <targetDir>/genesis.json
	fp, err := os.OpenFile(filepath.Clean(filepath.Join(f.targetDir, "genesis.json")), fileOpenflag, flieOpenMode)
	if err != nil {
		return nil, err
	}

	return fp, nil
}

// WriteRawJSON wtites the encoded JSON data to desinated target directory and the
// file.
func (f *FileGenesisTarget) WriteRawJSON(rawBz json.RawMessage) (rerr error) {
	if err := os.MkdirAll(f.targetDir, dirCreateMode); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", f.targetDir, err)
	}

	if len(f.moduleName) == 0 {
		return fmt.Errorf("failed to write RawJSON: empty module name")
	}

	fName := fmt.Sprintf("%s.json", f.moduleName)
	fPath := filepath.Join(f.targetDir, fName)
	fp, err := os.OpenFile(filepath.Clean(fPath), fileOpenflag, flieOpenMode)
	if err != nil {
		return fmt.Errorf("failed to create file, %s: %w", fPath, err)
	}

	defer func() {
		if err := fp.Close(); err != nil {
			if rerr != nil {
				rerr = fmt.Errorf("failed to close file %s: %s, %w", fName, err.Error(), rerr)
				return
			}

			rerr = fmt.Errorf("failed to close file %s: %w", fName, err)
		}
	}()

	if f.indent {
		rawBz, err = json.MarshalIndent(rawBz, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to format the raw JSON data: %w", err)
		}
	}

	n, err := fp.Write(rawBz)
	if err != nil {
		return fmt.Errorf("failed to write genesis file %s: %w", fName, err)
	}

	if n != len(rawBz) {
		return fmt.Errorf("failed to written %s, expect:%d, actual: %d", fName, len(rawBz), n)
	}

	return nil
}

// WriteMessage is an unsupported op.
func (f *FileGenesisTarget) WriteMessage(proto.Message) error {
	return errors.New("unsupported op")
}
