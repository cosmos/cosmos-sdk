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

	"cosmossdk.io/core/appmodule"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type FileGenesisSource struct {
	sourceDir  string
	moduleName string

	// the RawMessage from the genesis.json app_state.<module> that got passed into InitChain
	moduleRootJson json.RawMessage
	// flag for loading merged module states
	mergedStates bool
}

const (
	fileOpenflag  = os.O_CREATE | os.O_WRONLY
	flieOpenMode  = fs.FileMode(0o600)
	dirCreateMode = fs.FileMode(0o700)
)

// NewFileGenesisSource returns a new GenesisSource for the provided
// source directory and the provided module name where it is assumed
// that it contains encoded JSON data in the file or in the moduleState
// of the appState be passed from RequestInitChain.
func NewFileGenesisSource(sourceDir, moduleName string, rawModuleState json.RawMessage) GenesisSource {
	return &FileGenesisSource{
		sourceDir:      filepath.Clean(sourceDir),
		moduleName:     moduleName,
		moduleRootJson: rawModuleState,
	}
}

// OpenReader opens the source field reading from the given parameters,
// and returns a ReadCloser.
// It will try to open the field in order following by:
// <sourceDir>/<module>/<field>.json
// <field> key inside <sourceDir>/<module>.json
// app_state.<module>.<field> key from moduleRootJson
func (f *FileGenesisSource) OpenReader(field string) (io.ReadCloser, error) {
	// try reading genesis data from <sourceDir>/<module>/<field>.json
	fName := fmt.Sprintf("%s.json", field)
	fPath := filepath.Join(f.sourceDir, f.moduleName)

	fp, err := os.Open(filepath.Clean(filepath.Join(fPath, fName)))
	if err == nil {
		return fp, nil
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("unexpected error: %w", err)
	}

	if f.mergedStates {
		return f.unmarshalRawModuleWithField(f.moduleRootJson, field)
	}

	// try reading from <sourceDir>/<module>.json
	rawBz, err := f.ReadRawJSON()
	if err != nil {
		return nil, err
	}

	return f.unmarshalRawModuleWithField(rawBz, field)
}

func (f *FileGenesisSource) unmarshalRawModuleWithField(rawBz []byte, field string) (io.ReadCloser, error) {
	fieldState := make(map[string]json.RawMessage)
	err := json.Unmarshal(rawBz, &fieldState)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal fields from module state %s, err: %w", f.moduleName, err)
	}

	fieldRawData := fieldState[field]
	if fieldRawData == nil {
		return nil, fmt.Errorf("failed to retrieve module field %s/%s from genesis.json", f.moduleName, field)
	}

	// wrap raw field data to a ReadCloser
	return io.NopCloser(bytes.NewReader(fieldRawData)), nil
}

// ReadMessage reads rawJSON data from source file or moduleRawData,
// and then unmarshal it into proto.Message
func (f *FileGenesisSource) ReadMessage(msg proto.Message) error {
	bz, err := f.ReadRawJSON()
	if err != nil {
		return fmt.Errorf("unexpected error: %w", err)
	}

	return protojson.Unmarshal(bz, msg)
}

// ReadRawJSON returns a json.RawMessage read from the source file given by the
// source directory and the module name.
// Return the rawModuleJson coming from Initchain if the err is equal to ErrNotExist
func (f *FileGenesisSource) ReadRawJSON() (rawBz json.RawMessage, rerr error) {
	fName := fmt.Sprintf("%s.json", f.moduleName)
	fPath := filepath.Join(f.sourceDir, fName)

	fp, err := os.Open(filepath.Clean(fPath))
	if err != nil {
		if os.IsNotExist(err) {
			f.mergedStates = true
			return f.moduleRootJson, nil
		}
		return nil, err
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

	buf, err := io.ReadAll(fp)
	if err != nil {
		rerr = fmt.Errorf("failed to read file %s: %w", fp.Name(), err)
		return nil, rerr
	}
	if int64(len(buf)) != fi.Size() {
		rerr = fmt.Errorf("couldn't read entire file: %s, read: %d, file size: %d", fp.Name(), len(buf), fi.Size())
		return nil, rerr
	}

	// if the file is empty, return moduleRootJson instead
	if len(buf) == 0 {
		f.mergedStates = true
		return f.moduleRootJson, nil
	}

	// if moduleRootJson is empty, no data combined needed
	if len(f.moduleRootJson) == 0 {
		f.moduleRootJson = buf
		f.mergedStates = true
		return f.moduleRootJson, nil
	}

	// else, combine data with moduleRootJson
	moduleStates := make(map[string]interface{})
	if err := json.Unmarshal(f.moduleRootJson, &moduleStates); err != nil {
		rerr = fmt.Errorf("failed to unmarshal moduleRootJson: %w", err)
		return nil, rerr
	}

	if err := json.Unmarshal(buf, &moduleStates); err != nil {
		rerr = fmt.Errorf("failed to unmarshal the source module file: %w", err)
		return nil, rerr
	}

	bz, err := json.Marshal(moduleStates)
	if err != nil {
		rerr = fmt.Errorf("failed to marshal the combined module states: %w", err)
		return nil, rerr
	}

	// overwrite moduleRootJson with updated moduleStates
	f.moduleRootJson = bz
	f.mergedStates = true

	return f.moduleRootJson, nil
}

// SourceFromFile opens the source field reading from the given parameters,
// and returns appmodule.GenesisSource.
// It will try to open the field in order following by:
// <sourceDir>/<field>.json
func SourceFromFile(sourceDir string) (appmodule.GenesisSource, error) {
	return func(field string) (io.ReadCloser, error) {
		fName := fmt.Sprintf("%s.json", field)
		fPath := filepath.Clean(filepath.Join(sourceDir, fName))
		fp, err := os.Open(fPath)
		if err != nil {
			return nil, fmt.Errorf("failed to open the source file: %s, err: %w", fPath, err)
		}

		return fp, nil
	}, nil
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
		return os.OpenFile(filepath.Clean(filepath.Join(f.targetDir, f.moduleName, fileName)), fileOpenflag, flieOpenMode)
	}

	if err := os.MkdirAll(f.targetDir, dirCreateMode); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// if there is empty field, try to open/create a file to <targetDir>/<module>.json
	if len(f.moduleName) > 0 {
		fName := fmt.Sprintf("%s.json", f.moduleName)
		return os.OpenFile(filepath.Clean(filepath.Join(f.targetDir, fName)), fileOpenflag, flieOpenMode)
	}

	// else if there is empty module and field name try to open/create a file to <targetDir>/genesis.json
	return os.OpenFile(filepath.Clean(filepath.Join(f.targetDir, "genesis.json")), fileOpenflag, flieOpenMode)
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

// TargetToFile create a file for writing the genesus state to the file.
// It will try to create a file <targetDir>/<field>.json
func TargetToFile(targetDir string) appmodule.GenesisTarget {
	return func(field string) (io.WriteCloser, error) {
		if err := os.MkdirAll(targetDir, dirCreateMode); err != nil {
			return nil, fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
		}

		fName := fmt.Sprintf("%s.json", field)
		fPath := filepath.Clean(filepath.Join(targetDir, fName))
		return os.OpenFile(fPath, fileOpenflag, flieOpenMode)
	}
}
