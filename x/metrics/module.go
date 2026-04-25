package metrics

import (
	"fmt"
	"time"
)

const ModuleName = "metrics"

// AppModule implements the sdk.AppModule interface for the metrics module.
type AppModule struct {
	keeper     *Keeper
	blockStart time.Time
}

// NewAppModule creates a new AppModule instance.
func NewAppModule(keeper *Keeper) AppModule {
	return AppModule{keeper: keeper}
}

// Name returns the module name.
func (am AppModule) Name() string {
	return ModuleName
}

// RegisterServices is a no-op for the metrics module.
func (am AppModule) RegisterServices() {}

// BeginBlock records the start time for block processing measurement.
func (am *AppModule) BeginBlock() {
	am.blockStart = time.Now()
}

// EndBlock records the block processing duration.
func (am *AppModule) EndBlock() {
	if am.blockStart.IsZero() {
		return
	}
	duration := time.Since(am.blockStart)
	am.keeper.RecordBlockTime(duration)
}

// InitModule initializes the metrics module. It uses explicit error variable
// names to avoid shadowing outer errors during initialization.
func (am AppModule) InitModule() error {
	var initErr error

	if am.keeper == nil {
		return fmt.Errorf("keeper must not be nil")
	}

	if am.keeper.config.Namespace == "" {
		initErr = fmt.Errorf("namespace must not be empty")
		return initErr
	}

	if am.keeper.config.CollectionInterval <= 0 {
		initErr = fmt.Errorf("collection interval must be positive")
		return initErr
	}

	return nil
}
