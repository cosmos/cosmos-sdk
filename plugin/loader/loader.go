package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/spf13/cast"
	logging "github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/plugin"
	serverTypes "github.com/cosmos/cosmos-sdk/server/types"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
)

var preloadPlugins []plugin.Plugin

// Preload adds one or more plugins to the preload list. This should _only_ be called during init.
func Preload(plugins ...plugin.Plugin) {
	preloadPlugins = append(preloadPlugins, plugins...)
}

var loadPluginFunc = func(string) ([]plugin.Plugin, error) {
	return nil, fmt.Errorf("unsupported platform %s", runtime.GOOS)
}

type loaderState int

const (
	loaderLoading loaderState = iota
	loaderInitializing
	loaderInitialized
	loaderInjecting
	loaderInjected
	loaderStarting
	loaderStarted
	loaderClosing
	loaderClosed
	loaderFailed
)

func (ls loaderState) String() string {
	switch ls {
	case loaderLoading:
		return "Loading"
	case loaderInitializing:
		return "Initializing"
	case loaderInitialized:
		return "Initialized"
	case loaderInjecting:
		return "Injecting"
	case loaderInjected:
		return "Injected"
	case loaderStarting:
		return "Starting"
	case loaderStarted:
		return "Started"
	case loaderClosing:
		return "Closing"
	case loaderClosed:
		return "Closed"
	case loaderFailed:
		return "Failed"
	default:
		return "Unknown"
	}
}

// PluginLoader keeps track of loaded plugins.
//
// To use:
// 1. Load any desired plugins with Load and LoadDirectory. Preloaded plugins
//    will automatically be loaded.
// 2. Call Initialize to run all initialization logic.
// 3. Call Inject to register the plugins.
// 4. Optionally call Start to start plugins.
// 5. Call Close to close all plugins.
type PluginLoader struct {
	state    loaderState
	plugins  map[string]plugin.Plugin
	started  []plugin.Plugin
	opts     serverTypes.AppOptions
	logger   logging.Logger
	disabled []string
}

// NewPluginLoader creates new plugin loader
func NewPluginLoader(opts serverTypes.AppOptions, logger logging.Logger) (*PluginLoader, error) {
	loader := &PluginLoader{plugins: make(map[string]plugin.Plugin, len(preloadPlugins)), opts: opts, logger: logger}
	loader.disabled = cast.ToStringSlice(opts.Get(fmt.Sprintf("%s.%s", plugin.PLUGINS_TOML_KEY, plugin.PLUGINS_DISABLED_TOML_KEY)))
	for _, v := range preloadPlugins {
		if err := loader.Load(v); err != nil {
			return nil, err
		}
	}
	pluginDir := cast.ToString(opts.Get(plugin.PLUGINS_DIR_TOML_KEY))
	if pluginDir == "" {
		pluginDir = filepath.Join(os.Getenv("GOPATH"), plugin.DEFAULT_PLUGINS_DIRECTORY)
	}
	if err := loader.LoadDirectory(pluginDir); err != nil {
		return nil, err
	}
	return loader, nil
}

func (loader *PluginLoader) assertState(state loaderState) error {
	if loader.state != state {
		return fmt.Errorf("loader state must be %s, was %s", state, loader.state)
	}
	return nil
}

func (loader *PluginLoader) transition(from, to loaderState) error {
	if err := loader.assertState(from); err != nil {
		return err
	}
	loader.state = to
	return nil
}

// Load loads a plugin into the plugin loader.
func (loader *PluginLoader) Load(pl plugin.Plugin) error {
	if err := loader.assertState(loaderLoading); err != nil {
		return err
	}

	name := pl.Name()
	if ppl, ok := loader.plugins[name]; ok {
		// plugin is already loaded
		return fmt.Errorf(
			"plugin: %s, is duplicated in version: %s, "+
				"while trying to load dynamically: %s",
			name, ppl.Version(), pl.Version())
	}
	if sliceContainsStr(loader.disabled, name) {
		loader.logger.Info("not loading disabled plugin", "name", name)
		return nil
	}
	loader.plugins[name] = pl
	return nil
}

func sliceContainsStr(slice []string, str string) bool {
	for _, ele := range slice {
		if ele == str {
			return true
		}
	}
	return false
}

// LoadDirectory loads a directory of plugins into the plugin loader.
func (loader *PluginLoader) LoadDirectory(pluginDir string) error {
	if err := loader.assertState(loaderLoading); err != nil {
		return err
	}
	newPls, err := loader.loadDynamicPlugins(pluginDir)
	if err != nil {
		return err
	}

	for _, pl := range newPls {
		if err := loader.Load(pl); err != nil {
			return err
		}
	}
	return nil
}

func (loader *PluginLoader) loadDynamicPlugins(pluginDir string) ([]plugin.Plugin, error) {
	_, err := os.Stat(pluginDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var plugins []plugin.Plugin

	err = filepath.Walk(pluginDir, func(fi string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if fi != pluginDir {
				loader.logger.Info("found directory inside plugins directory", "directory", fi)
			}
			return nil
		}

		if info.Mode().Perm()&0111 == 0 {
			// file is not executable let's not load it
			// this is to prevent loading plugins from for example non-executable
			// mounts, some /tmp mounts are marked as such for security
			loader.logger.Error("non-executable file in plugins directory", "file", fi)
			return nil
		}

		if newPlugins, err := loadPluginFunc(fi); err == nil {
			plugins = append(plugins, newPlugins...)
		} else {
			return fmt.Errorf("loading plugin %s: %s", fi, err)
		}
		return nil
	})

	return plugins, err
}

// Initialize initializes all loaded plugins
func (loader *PluginLoader) Initialize() error {
	if err := loader.transition(loaderLoading, loaderInitializing); err != nil {
		return err
	}
	for name, p := range loader.plugins {
		if err := p.Init(loader.opts); err != nil {
			loader.state = loaderFailed
			return fmt.Errorf("unable to initialize plugin %s: %v", name, err)
		}
	}

	return loader.transition(loaderInitializing, loaderInitialized)
}

// Inject hooks all the plugins into the BaseApp.
func (loader *PluginLoader) Inject(bApp *baseapp.BaseApp, marshaller codec.BinaryCodec, keys map[string]*storeTypes.KVStoreKey) error {
	if err := loader.transition(loaderInitialized, loaderInjecting); err != nil {
		return err
	}

	for _, pl := range loader.plugins {
		if pl, ok := pl.(plugin.StateStreamingPlugin); ok {
			if err := pl.Register(bApp, marshaller, keys); err != nil {
				loader.state = loaderFailed
				return err
			}
		}
	}

	return loader.transition(loaderInjecting, loaderInjected)
}

// Start starts all long-running plugins.
func (loader *PluginLoader) Start(wg *sync.WaitGroup) error {
	if err := loader.transition(loaderInjected, loaderStarting); err != nil {
		return err
	}
	for _, pl := range loader.plugins {
		if pl, ok := pl.(plugin.StateStreamingPlugin); ok {
			if err := pl.Start(wg); err != nil {
				return err
			}
			loader.started = append(loader.started, pl)
		}
	}

	return loader.transition(loaderStarting, loaderStarted)
}

// Close stops all long-running plugins.
func (loader *PluginLoader) Close() error {
	switch loader.state {
	case loaderClosing, loaderFailed, loaderClosed:
		// nothing to do.
		return nil
	}
	loader.state = loaderClosing

	var errs []string
	started := loader.started
	loader.started = nil
	for _, pl := range started {
		if err := pl.Close(); err != nil {
			errs = append(errs, fmt.Sprintf(
				"error closing plugin %s: %s",
				pl.Name(),
				err.Error(),
			))
		}
	}
	if errs != nil {
		loader.state = loaderFailed
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	loader.state = loaderClosed
	return nil
}
