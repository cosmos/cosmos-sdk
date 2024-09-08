# LogMonitor

The `LogMonitor` is a component designed to monitor log outputs and trigger specific actions based on predefined patterns. It is particularly useful for detecting critical errors or specific log messages that require immediate attention, such as shutting down the application.

## Features

* **Log Monitoring**: Continuously monitors log outputs for specific patterns.
* **Shutdown Trigger**: Triggers a shutdown function when a predefined pattern is detected in the logs.

## Configuration

The `LogMonitor` can be configured via the `app.toml` configuration file. The relevant section in the configuration file is `[log-monitor]`.

### Example Configuration

```toml
[log-monitor]

# Enable defines if the log monitor should be enabled.

enabled = true

# ShutdownStrings defines the strings that will trigger a shutdown if found in the logs.

shutdown-strings = ["CONSENSUS FAILURE!", "CRITICAL ERROR"]
```

* `enabled`: A boolean value to enable or disable the log monitor.
* `shutdown-strings`: A list of strings that, if found in the logs, will trigger the shutdown function and terminate the application.

## Automatic Termination

One of the key features of the LogMonitor is its ability to automatically terminate the application when it detects a critical log message. 

When a log message matching any of the `shutdown-strings` is detected, the LogMonitor will:

1. Call the provided shutdown function.
2. Terminate the application process.

This feature ensures that the application can respond quickly to critical issues without requiring manual intervention.
