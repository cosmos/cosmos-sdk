# Log

The `cosmossdk.io/log` package provides a structured logging implementation for the Cosmos SDK using Go's standard `log/slog` with OpenTelemetry integration.

## Features

- **Dual output**: Logs are sent to both console and OpenTelemetry simultaneously
- **Pretty console output**: Human-readable colored output powered by [zerolog](https://github.com/rs/zerolog)
- **OpenTelemetry integration**: Full observability with trace correlation
- **Verbose mode**: Dynamic log level switching for operations like chain upgrades
- **Module filtering**: Filter logs by module and level

## Usage

### Basic Usage

```go
import "cosmossdk.io/log"

// Create a logger with default settings (console + OTEL)
logger := log.NewLogger("my-app")

// Log messages
logger.Info("server started", "port", 8080)
logger.Debug("processing request", "id", "abc123")
logger.Error("operation failed", "error", err)

// Add context to all subsequent logs
moduleLogger := logger.With("module", "auth")
moduleLogger.Info("user authenticated", "user_id", 42)
```

### With Context (Trace Correlation)

```go
// Use context-aware methods for trace/span correlation
logger.InfoContext(ctx, "handling request", "path", "/api/v1/users")
logger.ErrorContext(ctx, "request failed", "error", err)
```

### Configuration Options

```go
logger := log.NewLogger("my-app",
    // Set minimum log level
    log.WithLevel(slog.LevelDebug),
    
    // Enable verbose mode support (for upgrades, etc.)
    log.WithVerboseLevel(slog.LevelDebug),
    
    // Customize console output
    log.WithColor(true),                    // Enable colored output (default: true)
    log.WithTimeFormat(time.RFC3339),       // Custom time format (default: time.Kitchen)
    log.WithConsoleWriter(os.Stdout),       // Custom output writer (default: os.Stderr)
    
    // Output JSON instead of pretty text
    log.WithJSONOutput(),
    
    // Disable console (OTEL only)
    log.WithoutConsole(),
    
    // Custom OTEL provider
    log.WithLoggerProvider(provider),
    
    // Filter logs by module/level
    log.WithFilter(filterFunc),
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
- Console output respects level and filter settings
- OpenTelemetry receives all logs unfiltered (for full observability)
- Verbose mode only affects console output

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
    Impl() any  // Returns underlying *slog.Logger
}
```

