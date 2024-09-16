package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pelletier/go-toml/v2"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/log"
	"cosmossdk.io/tools/cosmovisor"
)

const (
	notset            = " is not set"
	cosmovisorDirName = "cosmovisor"

	cfgFileWithExt = "config.toml"
)

type InitTestSuite struct {
	suite.Suite
}

func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}

// cosmovisorInitEnv are some string values of environment variables used to configure Cosmovisor, and used by the init command.
type cosmovisorInitEnv struct {
	Home           string
	Name           string
	ColorLogs      string
	TimeFormatLogs string
}

type envMap struct {
	val        string
	allowEmpty bool
}

// ToMap creates a map of the cosmovisorInitEnv where the keys are the env var names.
func (c cosmovisorInitEnv) ToMap() map[string]envMap {
	return map[string]envMap{
		cosmovisor.EnvHome:           {val: c.Home, allowEmpty: false},
		cosmovisor.EnvName:           {val: c.Name, allowEmpty: false},
		cosmovisor.EnvColorLogs:      {val: c.ColorLogs, allowEmpty: false},
		cosmovisor.EnvTimeFormatLogs: {val: c.TimeFormatLogs, allowEmpty: true},
	}
}

// Set sets the field in this cosmovisorInitEnv corresponding to the provided envVar to the given envVal.
func (c *cosmovisorInitEnv) Set(envVar, envVal string) {
	switch envVar {
	case cosmovisor.EnvHome:
		c.Home = envVal
	case cosmovisor.EnvName:
		c.Name = envVal
	case cosmovisor.EnvColorLogs:
		c.Name = envVal
	case cosmovisor.EnvTimeFormatLogs:
		c.Name = envVal
	default:
		panic(fmt.Errorf("Unknown environment variable [%s]. Cannot set field to [%s]. ", envVar, envVal))
	}
}

// clearEnv clears environment variables and returns what they were.
// Designed to be used like this:
//
//	initialEnv := clearEnv()
//	defer setEnv(nil, initialEnv)
func (s *InitTestSuite) clearEnv() *cosmovisorInitEnv {
	s.T().Logf("Clearing environment variables.")
	rv := cosmovisorInitEnv{}
	for envVar := range rv.ToMap() {
		rv.Set(envVar, os.Getenv(envVar))
		s.Require().NoError(os.Unsetenv(envVar))
		viper.Reset()
	}
	return &rv
}

// setEnv sets environment variables to the values provided.
// If t is not nil, and there's a problem, the test will fail immediately.
// If t is nil, problems will just be logged using s.T().
func (s *InitTestSuite) setEnv(t *testing.T, env *cosmovisorInitEnv) { //nolint:thelper // false psotive
	if t == nil {
		s.T().Logf("Restoring environment variables.")
	}
	for envVar, envVal := range env.ToMap() {
		var err error
		var msg string
		if len(envVal.val) != 0 || envVal.allowEmpty {
			err = os.Setenv(envVar, envVal.val)
			msg = fmt.Sprintf("setting %s to %s", envVar, envVal.val)
		} else {
			err = os.Unsetenv(envVar)
			msg = fmt.Sprintf("unsetting %s", envVar)
		}
		switch {
		case t != nil:
			require.NoError(t, err, msg)
		case err != nil:
			s.T().Logf("error %s: %v", msg, err)
		default:
			s.T().Logf("done %s", msg)
		}
	}
}

// readStdInpFromFile reads the provided data as if it were a standard input.
func (s *InitTestSuite) readStdInpFromFile(data []byte) {
	// Create a temporary file and write the test input into it
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		s.T().Fatal(err)
	}

	// write the test input into the temporary file
	if _, err := tmpfile.Write(data); err != nil {
		s.T().Fatal(err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		s.T().Fatal(err)
	}

	os.Stdin = tmpfile
}

var (
	_ io.Reader = BufferedPipe{}
	_ io.Writer = BufferedPipe{}
)

// BufferedPipe contains a connected read/write pair of files (a pipe),
// and a buffer of what goes through it that is populated in the background.
type BufferedPipe struct {
	// Name is a string to help humans identify this BufferedPipe.
	Name string
	// Reader is the reader end of the pipe.
	Reader *os.File
	// Writer is the writer end of the pipe.
	Writer *os.File
	// BufferReader is the reader used by this BufferedPipe while buffering.
	// If this BufferedPipe is not replicating to anything, it will be the same as the Reader.
	// Otherwise, it will be a reader encapsulating all desired replication.
	BufferReader io.Reader
	// Error is the last error encountered by this BufferedPipe.
	Error error

	// buffer is the channel used to communicate buffer contents.
	buffer chan []byte
	// stated is true if this BufferedPipe has been started.
	started bool
}

// NewBufferedPipe creates a new BufferedPipe with the given name.
// Files must be closed once you are done with them (e.g. with .Close()).
// Once ready, buffering must be started using .Start(). See also StartNewBufferedPipe.
func NewBufferedPipe(name string, replicateTo ...io.Writer) (BufferedPipe, error) {
	p := BufferedPipe{Name: name}
	p.Reader, p.Writer, p.Error = os.Pipe()
	if p.Error != nil {
		return p, p.Error
	}
	p.BufferReader = p.Reader
	p.AddReplicationTo(replicateTo...)
	return p, nil
}

// StartNewBufferedPipe creates a new BufferedPipe and starts it.
//
// This is functionally equivalent to:
//
//	p, _ := NewBufferedPipe(name, replicateTo...)
//	p.Start()
func StartNewBufferedPipe(name string, replicateTo ...io.Writer) (BufferedPipe, error) {
	p, err := NewBufferedPipe(name, replicateTo...)
	if err != nil {
		return p, err
	}
	p.Start()
	return p, nil
}

// AddReplicationTo adds replication of this buffered pipe to the provided writers.
//
// Panics if this BufferedPipe is already started.
func (p *BufferedPipe) AddReplicationTo(writers ...io.Writer) {
	p.panicIfStarted("cannot add further replication")
	for _, writer := range writers {
		p.BufferReader = io.TeeReader(p.BufferReader, writer)
	}
}

// Start initiates buffering in a background process.
//
// Panics if this BufferedPipe is already started.
func (p *BufferedPipe) Start() {
	p.panicIfStarted("cannot restart")
	p.buffer = make(chan []byte)
	go func() {
		var b bytes.Buffer
		if _, p.Error = io.Copy(&b, p.BufferReader); p.Error != nil {
			b.WriteString("buffer error: " + p.Error.Error())
		}
		p.buffer <- b.Bytes()
	}()
	p.started = true
}

// IsStarted returns true if this BufferedPipe has already been started.
func (p *BufferedPipe) IsStarted() bool {
	return p.started
}

// IsBuffering returns true if this BufferedPipe has started buffering and has not yet been collected.
func (p *BufferedPipe) IsBuffering() bool {
	return p.buffer != nil
}

// Collect closes this pipe's writer then blocks, returning with the final buffer contents once available.
// If Collect() has previously been called on this BufferedPipe, an empty byte slice is returned.
//
// Panics if this BufferedPipe has not been started.
func (p *BufferedPipe) Collect() []byte {
	if !p.started {
		panic("buffered pipe " + p.Name + " has not been started: cannot collect")
	}
	_ = p.Writer.Close()
	if p.buffer == nil {
		return []byte{}
	}
	rv := <-p.buffer
	p.buffer = nil
	return rv
}

// Read implements the io.Reader interface on this BufferedPipe.
func (p BufferedPipe) Read(bz []byte) (n int, err error) {
	return p.Reader.Read(bz)
}

// Write implements the io.Writer interface on this BufferedPipe.
func (p BufferedPipe) Write(bz []byte) (n int, err error) {
	return p.Writer.Write(bz)
}

// Close makes sure the files in this BufferedPipe are closed.
func (p *BufferedPipe) Close() {
	_ = p.Reader.Close()
	_ = p.Writer.Close()
}

// panicIfStarted panics if this BufferedPipe has been started.
func (p *BufferedPipe) panicIfStarted(msg string) {
	if p.started {
		panic("buffered pipe " + p.Name + " already started: " + msg)
	}
}

// NewCapturingLogger creates a buffered stdout pipe, and a logger that uses it.
func (s *InitTestSuite) NewCapturingLogger() (*BufferedPipe, log.Logger) {
	bufferedStdOut, err := StartNewBufferedPipe("stdout", os.Stdout)
	s.Require().NoError(err, "creating stdout buffered pipe")
	logger := log.NewLogger(bufferedStdOut, log.ColorOption(false), log.TimeFormatOption(time.RFC3339Nano)).With(log.ModuleKey, cosmovisorDirName)
	return &bufferedStdOut, logger
}

// CreateHelloWorld creates a shell script that outputs HELLO WORLD.
// It will have the provided filemode and be in a freshly made temp directory.
// The returned string is the full path to the new file.
func (s *InitTestSuite) CreateHelloWorld(filemode os.FileMode) string {
	tmpDir := s.T().TempDir()
	tmpExe := filepath.Join(tmpDir, "hello-world.sh")
	tmpExeBz := []byte(`#!/bin/sh
echo 'HELLO WORLD'
`)
	s.Require().NoError(os.WriteFile(tmpExe, tmpExeBz, filemode))
	return tmpExe
}

func (s *InitTestSuite) TestInitializeCosmovisorNegativeValidation() {
	initEnv := s.clearEnv()
	defer s.setEnv(nil, initEnv)

	tmpExe := s.CreateHelloWorld(0o755)

	tmpDir := s.T().TempDir()

	tests := []struct {
		name  string
		env   cosmovisorInitEnv
		args  []string
		inErr []string
	}{
		{
			name:  "no args",
			env:   cosmovisorInitEnv{Home: "/example", Name: "foo"},
			args:  []string{},
			inErr: []string{"no <path to executable> provided"},
		},
		{
			name:  "one empty arg",
			env:   cosmovisorInitEnv{Home: "/example", Name: "foo"},
			args:  []string{""},
			inErr: []string{"no <path to executable> provided"},
		},
		{
			name:  "exe not found",
			env:   cosmovisorInitEnv{Home: "/example", Name: "foo"},
			args:  []string{filepath.Join(tmpDir, "not-gonna-find-me")},
			inErr: []string{"executable file not found", "not-gonna-find-me"},
		},
		{
			name:  "exe is a dir",
			env:   cosmovisorInitEnv{Home: "/example", Name: "foo"},
			args:  []string{tmpDir},
			inErr: []string{"invalid path to executable: must not be a directory"},
		},
		{
			name:  "no name",
			env:   cosmovisorInitEnv{Home: "/example", Name: ""},
			args:  []string{tmpExe},
			inErr: []string{cosmovisor.EnvName + notset},
		},
		{
			name:  "no home",
			env:   cosmovisorInitEnv{Home: "", Name: "foo"},
			args:  []string{tmpExe},
			inErr: []string{cosmovisor.EnvHome + notset},
		},
		{
			name:  "home is relative",
			env:   cosmovisorInitEnv{Home: "./home", Name: "foo"},
			args:  []string{tmpExe},
			inErr: []string{cosmovisor.EnvHome + " must be an absolute path"},
		},
		{
			name:  "no name and no home",
			env:   cosmovisorInitEnv{Home: "", Name: ""},
			args:  []string{tmpExe},
			inErr: []string{cosmovisor.EnvName + notset, cosmovisor.EnvHome + notset},
		},
	}

	for _, tc := range tests {
		s.T().Run(tc.name, func(t *testing.T) {
			s.setEnv(t, &tc.env)
			buffer, logger := s.NewCapturingLogger()
			err := InitializeCosmovisor(logger, tc.args)
			require.Error(t, err)
			for _, exp := range tc.inErr {
				require.ErrorContains(t, err, exp)
			}
			// And make sure there wasn't any log output.
			// Log output indicates that work is being done despite validation errors.
			outputBz := buffer.Collect()
			outputStr := string(outputBz)
			require.Equal(t, "", outputStr, "log output")
		})
	}
}

func (s *InitTestSuite) TestInitializeCosmovisorInvalidExisting() {
	initEnv := s.clearEnv()
	defer s.setEnv(nil, initEnv)

	hwExe := s.CreateHelloWorld(0o755)

	s.T().Run("genesis bin is not a directory", func(t *testing.T) {
		testDir := t.TempDir()
		env := &cosmovisorInitEnv{
			Home: filepath.Join(testDir, "home"),
			Name: "pear",
		}
		genDir := filepath.Join(env.Home, cosmovisorDirName, "genesis")
		genBin := filepath.Join(genDir, "bin")
		require.NoError(t, os.MkdirAll(genDir, 0o755), "creating genesis directory")
		require.NoError(t, copyFile(hwExe, genBin), "copying exe to genesis/bin")

		s.setEnv(t, env)
		logger := log.NewNopLogger()
		expErr := fmt.Sprintf("the path %q already exists but is not a directory", genBin)
		err := InitializeCosmovisor(logger, []string{hwExe})
		require.EqualError(t, err, expErr, "invalid path to executable: must not be a directory", "calling InitializeCosmovisor")
	})

	s.T().Run("the EnsureBinary test fails", func(t *testing.T) {
		testDir := t.TempDir()
		env := &cosmovisorInitEnv{
			Home: filepath.Join(testDir, "home"),
			Name: "grapes",
		}
		// Create the genesis bin executable path fully as a directory (instead of a file).
		// That should get through all the other stuff, but error when EnsureBinary is called.
		genBinExe := filepath.Join(env.Home, cosmovisorDirName, "genesis", "bin", env.Name)
		require.NoError(t, os.MkdirAll(genBinExe, 0o755))
		expErr := fmt.Sprintf("%s is not a regular file", env.Name)
		// Check the log messages just to make sure it's erroring where expecting.
		expInLog := []string{
			"checking on the genesis/bin directory",
			"checking on the genesis/bin executable",
			fmt.Sprintf("the %q file already exists", genBinExe),
			fmt.Sprintf("making sure %q is executable", genBinExe),
		}
		expNotInLog := []string{
			"checking on the current symlink and creating it if needed",
			"the current symlink points to",
		}

		s.setEnv(t, env)
		buffer, logger := s.NewCapturingLogger()
		logger.Info(fmt.Sprintf("Calling InitializeCosmovisor: %s", t.Name()))
		err := InitializeCosmovisor(logger, []string{hwExe})
		require.EqualError(t, err, expErr, "calling InitializeCosmovisor")
		bufferBz := buffer.Collect()
		bufferStr := string(bufferBz)
		for _, exp := range expInLog {
			assert.Contains(t, bufferStr, exp, "expected log statement")
		}
		for _, notExp := range expNotInLog {
			assert.NotContains(t, bufferStr, notExp, "unexpected log statement")
		}
	})

	s.T().Run("current already exists as a file", func(t *testing.T) {
		testDir := t.TempDir()
		env := &cosmovisorInitEnv{
			Home: filepath.Join(testDir, "home"),
			Name: "orange",
		}
		rootDir := filepath.Join(env.Home, cosmovisorDirName)
		require.NoError(t, os.MkdirAll(rootDir, 0o755))
		curLn := filepath.Join(rootDir, "current")
		genDir := filepath.Join(rootDir, "genesis")
		require.NoError(t, copyFile(hwExe, curLn))
		expErr := fmt.Sprintf("symlink %s %s: file exists", genDir, curLn)

		s.setEnv(t, env)
		buffer, logger := s.NewCapturingLogger()
		logger.Info(fmt.Sprintf("Calling InitializeCosmovisor: %s", t.Name()))
		err := InitializeCosmovisor(logger, []string{hwExe})
		require.EqualError(t, err, expErr, "calling InitializeCosmovisor")
		bufferBz := buffer.Collect()
		bufferStr := string(bufferBz)
		assert.Contains(t, bufferStr, "checking on the current symlink and creating it if needed")
	})

	// Failure cases not tested:
	//	Cannot create genesis bin directory
	//		I had a test for this that created the `genesis` directory with permissions 0o555.
	//		I also tried it where it would create the directory at the root of the file system.
	//		In both cases, the test worked as expected locally, but not on the github runners. So it was removed.
	//	Given executable is not readable
	//		I had a test for this that created the executable with permissions 0o311.
	//		The test worked as expected locally, but not on the github runners. So it was removed.
	//	Cannot get info on the genesis bin directory.
	//		Not sure how to create a thing that will return
	//		an error other than a NotExists error when stat is called on it.
	//	Cannot write to genesis bin dir
	//		I had a test for this that created the bin dir with permissions 0o555.
	//		The test worked as expected locally, but not on the github runners. So it was removed.
	//	Cannot make the copied file executable.
	//		Probably need another user for this.
	//		Create the genesis bin file first, using the other user, and set permissions to 600.
}

func (s *InitTestSuite) TestInitializeCosmovisorValid() {
	initEnv := s.clearEnv()
	defer s.setEnv(nil, initEnv)

	hwNonExe := s.CreateHelloWorld(0o644)
	hwExe := s.CreateHelloWorld(0o755)

	s.T().Run("starting with blank slate", func(t *testing.T) {
		testDir := s.T().TempDir()
		env := &cosmovisorInitEnv{
			Home: filepath.Join(testDir, "home"),
			Name: "blank",
		}
		curLn := filepath.Join(env.Home, cosmovisorDirName, "current")
		genBinDir := filepath.Join(env.Home, cosmovisorDirName, "genesis", "bin")
		genBinExe := filepath.Join(genBinDir, env.Name)
		expInLog := []string{
			"checking on the genesis/bin directory",
			fmt.Sprintf("creating directory (and any parents): %q", genBinDir),
			"checking on the genesis/bin executable",
			fmt.Sprintf("copying executable into place: %q", genBinExe),
			fmt.Sprintf("making sure %q is executable", genBinExe),
			"checking on the current symlink and creating it if needed",
			fmt.Sprintf("the current symlink points to: %q", genBinExe),
			fmt.Sprintf("cosmovisor config.toml created at: %s", filepath.Join(env.Home, cosmovisorDirName, cfgFileWithExt)),
		}

		s.setEnv(s.T(), env)
		buffer, logger := s.NewCapturingLogger()
		logger.Info(fmt.Sprintf("Calling InitializeCosmovisor: %s", t.Name()))
		err := InitializeCosmovisor(logger, []string{hwNonExe})
		require.NoError(t, err, "calling InitializeCosmovisor")

		_, err = os.Stat(genBinDir)
		assert.NoErrorf(t, err, "statting the genesis bin dir: %q", genBinDir)
		_, err = os.Stat(curLn)
		assert.NoError(t, err, "statting the current link: %q", curLn)
		exeInfo, exeErr := os.Stat(genBinExe)
		if assert.NoError(t, exeErr, "statting the executable: %q", genBinExe) {
			assert.True(t, exeInfo.Mode().IsRegular(), "executable is regular file")
			// Check if the world-executable bit is set.
			exePermMask := exeInfo.Mode().Perm() & 0o001
			assert.NotEqual(t, 0, exePermMask, "executable mask")
		}
		bufferBz := buffer.Collect()
		bufferStr := string(bufferBz)
		for _, exp := range expInLog {
			assert.Contains(t, bufferStr, exp)
		}
	})

	s.T().Run("genesis and upgrades exist but no current", func(t *testing.T) {
		testDir := s.T().TempDir()
		env := &cosmovisorInitEnv{
			Home: filepath.Join(testDir, "home"),
			Name: "nocur",
		}
		rootDir := filepath.Join(env.Home, cosmovisorDirName)
		genBinDir := filepath.Join(rootDir, "genesis", "bin")
		genBinDirExe := filepath.Join(genBinDir, env.Name)
		require.NoError(t, os.MkdirAll(genBinDir, 0o755), "making genesis bin dir")
		require.NoError(t, copyFile(hwExe, genBinDirExe), "copying executable to genesis")
		upgradesDir := filepath.Join(rootDir, "upgrades")
		for i := 1; i <= 5; i++ {
			upgradeBinDir := filepath.Join(upgradesDir, fmt.Sprintf("upgrade-%02d", i), "bin")
			upgradeBinDirExe := filepath.Join(upgradeBinDir, env.Name)
			require.NoErrorf(t, os.MkdirAll(upgradeBinDir, 0o755), "Making upgrade %d bin dir", i)
			require.NoErrorf(t, copyFile(hwExe, upgradeBinDirExe), "copying executable to upgrade %d", i)
		}

		expInLog := []string{
			"checking on the genesis/bin directory",
			fmt.Sprintf("the %q directory already exists", genBinDir),
			"checking on the genesis/bin executable",
			fmt.Sprintf("the %q file already exists", genBinDirExe),
			fmt.Sprintf("making sure %q is executable", genBinDirExe),
			fmt.Sprintf("the current symlink points to: %q", genBinDirExe),
			fmt.Sprintf("cosmovisor config.toml created at: %s", filepath.Join(env.Home, cosmovisorDirName, cfgFileWithExt)),
		}

		s.setEnv(t, env)
		buffer, logger := s.NewCapturingLogger()
		logger.Info(fmt.Sprintf("Calling InitializeCosmovisor: %s", t.Name()))
		err := InitializeCosmovisor(logger, []string{hwExe})
		require.NoError(t, err, "calling InitializeCosmovisor")
		bufferBz := buffer.Collect()
		bufferStr := string(bufferBz)
		for _, exp := range expInLog {
			assert.Contains(t, bufferStr, exp)
		}
	})

	s.T().Run("genesis bin dir exists empty", func(t *testing.T) {
		testDir := s.T().TempDir()
		env := &cosmovisorInitEnv{
			Home: filepath.Join(testDir, "home"),
			Name: "emptygen",
		}
		rootDir := filepath.Join(env.Home, cosmovisorDirName)
		genBinDir := filepath.Join(rootDir, "genesis", "bin")
		genBinExe := filepath.Join(genBinDir, env.Name)
		require.NoError(t, os.MkdirAll(genBinDir, 0o755), "making genesis bin dir")

		expInLog := []string{
			"checking on the genesis/bin directory",
			fmt.Sprintf("the %q directory already exists", genBinDir),
			"checking on the genesis/bin executable",
			fmt.Sprintf("copying executable into place: %q", genBinExe),
			fmt.Sprintf("making sure %q is executable", genBinExe),
			fmt.Sprintf("the current symlink points to: %q", genBinExe),
			fmt.Sprintf("cosmovisor config.toml created at: %s", filepath.Join(env.Home, cosmovisorDirName, cfgFileWithExt)),
		}

		s.setEnv(t, env)
		buffer, logger := s.NewCapturingLogger()
		logger.Info(fmt.Sprintf("Calling InitializeCosmovisor: %s", t.Name()))
		err := InitializeCosmovisor(logger, []string{hwExe})
		require.NoError(t, err, "calling InitializeCosmovisor")
		bufferBz := buffer.Collect()
		bufferStr := string(bufferBz)
		for _, exp := range expInLog {
			assert.Contains(t, bufferStr, exp)
		}
	})

	s.T().Run("ask to override (y/n) the existing config file", func(t *testing.T) {
	})

	s.T().Run("init command exports configs to default path", func(t *testing.T) {
		testDir := s.T().TempDir()
		env := &cosmovisorInitEnv{
			Home: filepath.Join(testDir, "home"),
			Name: "emptygen",
		}

		s.setEnv(t, env)
		buffer, logger := s.NewCapturingLogger()
		logger.Info(fmt.Sprintf("Calling InitializeCosmovisor: %s", t.Name()))
		err := InitializeCosmovisor(logger, []string{hwExe})
		require.NoError(t, err, "calling InitializeCosmovisor")
		bufferBz := buffer.Collect()
		bufferStr := string(bufferBz)
		assert.Contains(t, bufferStr, fmt.Sprintf("cosmovisor config.toml created at: %s", filepath.Join(env.Home, cosmovisorDirName, cfgFileWithExt)))
	})
}

func (s *InitTestSuite) TestInitializeCosmovisorWithOverrideCfg() {
	initEnv := s.clearEnv()
	defer s.setEnv(nil, initEnv)

	tmpExe := s.CreateHelloWorld(0o755)
	testDir := s.T().TempDir()
	homePath := filepath.Join(testDir, "backup")
	testCases := []struct {
		name     string
		input    string
		cfg      *cosmovisor.Config
		override bool
	}{
		{
			name:  "yes override",
			input: "y\n",
			cfg: &cosmovisor.Config{
				Home:           homePath,
				Name:           "old_test",
				DataBackupPath: homePath,
			},
			override: true,
		},
		{
			name:  "no override",
			input: "n\n",
			cfg: &cosmovisor.Config{
				Home:           homePath,
				Name:           "old_test",
				DataBackupPath: homePath,
			},
			override: false,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			// create a root cosmovisor directory
			require.NoError(t, os.MkdirAll(tc.cfg.Root(), 0o755), "making root dir")

			// create a config file in the default location
			file, err := os.Create(tc.cfg.DefaultCfgPath())
			require.NoError(t, err)

			// write the config to the file
			err = toml.NewEncoder(file).Encode(tc.cfg)
			require.NoError(t, err)

			err = file.Close()
			require.NoError(t, err)

			s.readStdInpFromFile([]byte(tc.input))

			_, logger := s.NewCapturingLogger()
			logger.Info(fmt.Sprintf("Calling InitializeCosmovisor: %s", t.Name()))

			// override the daemon name in environment file
			// if override is true (y), then the name should be updated in the config file
			// otherwise (n), the name should not be updated in the config file
			s.setEnv(t, &cosmovisorInitEnv{
				Home: tc.cfg.Home,
				Name: "update_name",
			})

			err = InitializeCosmovisor(logger, []string{tmpExe})
			require.NoError(t, err, "calling InitializeCosmovisor")

			cfg := &cosmovisor.Config{}
			// read the config file
			cfgFile, err := os.Open(tc.cfg.DefaultCfgPath())
			require.NoError(t, err)
			defer cfgFile.Close()

			err = toml.NewDecoder(cfgFile).Decode(cfg)
			require.NoError(t, err)
			if tc.override {
				// check if the name is updated
				// basically, override the existing config file
				assert.Equal(t, "update_name", cfg.Name)
			} else {
				// daemon name should not be updated
				assert.Equal(t, tc.cfg.Name, cfg.Name)
			}
		})
	}
}
