# Log v2

The `cosmossdk.io/log/v2/v2` package provides a structured logging implementation for the Cosmos SDK using [zerolog](https://github.com/rs/zerolog) with optional OpenTelemetry integration.

## Features

- **Fast default path**: Zero-allocation logging via zerolog (OTEL disabled by default)
- **Pretty console output**: Human-readable colored output
- **Optional OpenTelemetry**: Enable with `WithOTEL()` for dual output (console + OTEL). This will happen automatically if an OpenTelemetry configuration is set with a logger provider.
- **Verbose mode**: Dynamic log level switching for operations like chain upgrades
- **Module filtering**: Filter logs by module and level
- **Stack traces**: Optional stack trace logging on errors

## Usage

### Basic Usage

```go
import "cosmossdk.io/log/v2/v2"

// Create a logger (fast zerolog path, no OTEL)
logger := log.NewLogger("my-app")

// Log messages
logger.Info("server started", "port", 8080)
logger.Debug("processing request", "id", "abc123")
logger.Error("operation failed", "error", err)

// Add context to all subsequent logs
moduleLogger := logger.With("module", "auth")
moduleLogger.Info("user authenticated", "user_id", 42)
```

### With OpenTelemetry

```go
// Enable OTEL for dual output (console + OpenTelemetry)
logger := log.NewLogger("my-app", log.WithOTEL())

// Use context-aware methods for trace/span correlation
logger.InfoContext(ctx, "handling request", "path", "/api/v1/users")
logger.ErrorContext(ctx, "request failed", "error", err)

// OTEL-only (no console output)
logger := log.NewLogger("my-app", log.WithOTEL(), log.WithoutConsole())

// Custom OTEL provider
logger := log.NewLogger("my-app", log.WithLoggerProvider(provider))
```

### Configuration Options

```go
logger := log.NewLogger("my-app",
    // Set minimum log level (default: slog.LevelInfo)
    log.WithLevel(slog.LevelDebug),
    
    // Enable verbose mode support (for upgrades, etc.)
    log.WithVerboseLevel(slog.LevelDebug),
    
    // Console formatting
    log.WithColor(true),                    // Enable colored output (default: true)
    log.WithTimeFormat(time.RFC3339),       // Custom time format (default: time.Kitchen)
    log.WithConsoleWriter(os.Stdout),       // Custom output writer (default: os.Stderr)
    log.WithJSONOutput(),                   // Output JSON instead of pretty text
    
    // OpenTelemetry
    log.WithOTEL(),                         // Enable OTEL forwarding
    log.WithoutOTEL(),                      // Explicitly disable OTEL
    log.WithLoggerProvider(provider),       // Custom OTEL provider (implies WithOTEL)
    log.WithoutConsole(),                   // Disable console (OTEL only)
    
    // Advanced
    log.WithFilter(filterFunc),             // Filter logs by module/level
    log.WithStackTrace(true),               // Enable stack traces on errors
    log.WithHooks(hook1, hook2),            // Add zerolog hooks
)
```

### Verbose Mode

Verbose mode allows dynamic log level switching, useful during chain upgrades:

```go
logger := log.NewLogger("cosmos-sdk",
    log.WithLevel(slog.LevelInfo),
    log.WithVerboseLevel(slog.LevelDebug),
    log.WithFilter(myFilter),
)

// Cast to VerboseModeLogger to access SetVerboseMode
if vl, ok := logger.(log.VerboseModeLogger); ok {
    // Enable verbose mode - lowers level to Debug and disables filter
    vl.SetVerboseMode(true)
    
    // ... perform upgrade ...
    
    // Disable verbose mode - restores original level and filter
    vl.SetVerboseMode(false)
}
```

### Log Level Filtering

Parse log level configuration strings:

```go
// Format: "module:level,module:level,..." or just "level" for default
filterFunc, err := log.ParseLogLevel("info,consensus:debug,p2p:error")

logger := log.NewLogger("my-app",
    log.WithFilter(filterFunc),
)
```

### Custom Logger

Wrap an existing `*slog.Logger`:

```go
slogger := slog.New(myCustomHandler)
logger := log.NewCustomLogger(slogger)
```

### No-Op Logger

For testing or disabled logging:

```go
logger := log.NewNopLogger()
```

## Architecture

The package supports two logging paths:

### Default Path (OTEL Disabled)

```
┌─────────────────────────────────────────────────────┐
│                  zeroLogWrapper                     │
│              (fast path, zero allocs)               │
├─────────────────────────────────────────────────────┤
│                    zerolog                          │
│  • Pretty or JSON formatting                        │
│  • Color support                                    │
│  • Level filtering                                  │
│  • Module filtering                                 │
│  • Verbose mode                                     │
│  • Stack traces                                     │
└─────────────────────────────────────────────────────┘
```

### OTEL Enabled Path

```
┌─────────────────────────────────────────────────────┐
│                    slog.Logger                      │
├─────────────────────────────────────────────────────┤
│                   multiHandler                      │
├────────────────────────┬────────────────────────────┤
│    zerologHandler      │       otelHandler          │
│  (console output)      │   (OpenTelemetry export)   │
│                        │                            │
│  • Pretty formatting   │  • Gets ALL logs           │
│  • Color support       │  • No filtering            │
│  • Level filtering     │  • Trace correlation       │
│  • Module filtering    │                            │
│  • Verbose mode        │                            │
└────────────────────────┴────────────────────────────┘
```

**Key design decisions:**
- OTEL is only enabled if an OpenTelemetry configuration is set for telemetry.
- When OTEL is enabled, console output respects level and filter settings
- OpenTelemetry receives all logs unfiltered (for full observability)
- Verbose mode only affects console output
- Context methods (`InfoContext`, etc.) enable trace correlation when OTEL is active

## Logger Interface

```go
type Logger interface {
    Info(msg string, keyVals ...any)
    Warn(msg string, keyVals ...any)
    Error(msg string, keyVals ...any)
    Debug(msg string, keyVals ...any)
    
    // Context-aware methods for trace correlation
    InfoContext(ctx context.Context, msg string, keyVals ...any)
    WarnContext(ctx context.Context, msg string, keyVals ...any)
    ErrorContext(ctx context.Context, msg string, keyVals ...any)
    DebugContext(ctx context.Context, msg string, keyVals ...any)
    
    With(keyVals ...any) Logger
    Impl() any  // Returns underlying *zerolog.Logger or *slog.Logger
}

type VerboseModeLogger interface {
    Logger
    SetVerboseMode(bool)
}
```
