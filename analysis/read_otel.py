"""Load OpenTelemetry JSONL data into DuckDB views."""

import duckdb
from pathlib import Path
from typing import Dict


def load_otel_data(data_path: str | Path) -> duckdb.DuckDBPyConnection:
    """
    Load OpenTelemetry data from a directory containing trace.jsonl, logs.jsonl, and metrics.jsonl.

    Creates three views in the returned DuckDB connection:
    - spans: flattened trace data
    - logs: flattened log data
    - metrics: flattened metrics data

    Args:
        data_path: Path to directory containing the JSONL files

    Returns:
        DuckDB connection with views created

    Example:
        >>> con = load_otel_data('/path/to/data')
        >>> con.sql("SELECT span_name, count(*) FROM spans GROUP BY span_name").show()
    """
    data_path = Path(data_path)
    con = duckdb.connect(':memory:')

    # Create spans view from trace.jsonl
    trace_file = data_path / 'trace.jsonl'
    if trace_file.exists():
        con.execute(f"""
            CREATE VIEW spans AS
            SELECT
                Name AS span_name,
                SpanContext.TraceID AS trace_id,
                SpanContext.SpanID AS span_id,
                Parent.SpanID AS parent_span_id,
                CAST(StartTime AS TIMESTAMPTZ) AS start_time,
                CAST(EndTime AS TIMESTAMPTZ) AS end_time,
                CAST(EndTime AS TIMESTAMPTZ) - CAST(StartTime AS TIMESTAMPTZ) AS duration,
                InstrumentationScope.Name AS scope,
                ChildSpanCount AS child_span_count,
                Attributes,
                Resource,
                Status
            FROM read_ndjson_auto('{trace_file}')
        """)

    # Create logs view from logs.jsonl
    logs_file = data_path / 'logs.jsonl'
    if logs_file.exists():
        con.execute(f"""
            CREATE VIEW logs AS
            SELECT
                CAST(Timestamp AS TIMESTAMPTZ) AS timestamp,
                CAST(ObservedTimestamp AS TIMESTAMPTZ) AS observed_timestamp,
                Severity AS severity,
                SeverityText AS severity_text,
                Body.Value AS body,
                TraceID AS trace_id,
                SpanID AS span_id,
                Attributes,
                Resource,
                Scope
            FROM read_ndjson_auto('{logs_file}')
        """)

    # Create metrics view from metrics.jsonl (if file exists and has data)
    metrics_file = data_path / 'metrics.jsonl'
    if metrics_file.exists() and metrics_file.stat().st_size > 0:
        # Flatten the deeply nested metrics structure
        # Structure: Resource → ScopeMetrics[] → Metrics[] → Data → DataPoints[]
        con.execute(f"""
            CREATE VIEW metrics AS
            SELECT
                scope_metric.Scope.Name AS scope_name,
                scope_metric.Scope.Version AS scope_version,
                metric.Name AS metric_name,
                metric.Description AS metric_description,
                metric.Unit AS unit,
                metric.Data.Temporality AS temporality,
                metric.Data.IsMonotonic AS is_monotonic,
                CAST(dp.Time AS TIMESTAMPTZ) AS time,
                CAST(dp.StartTime AS TIMESTAMPTZ) AS start_time,
                dp.Value AS value,
                dp.Count AS count,
                dp.Sum AS sum,
                dp.Min AS min,
                dp.Max AS max,
                dp.Attributes AS attributes,
                raw.Resource AS resource
            FROM read_ndjson_auto('{metrics_file}') AS raw
            CROSS JOIN UNNEST(raw.ScopeMetrics) AS t(scope_metric)
            CROSS JOIN UNNEST(scope_metric.Metrics) AS t2(metric)
            CROSS JOIN UNNEST(metric.Data.DataPoints) AS t3(dp)
        """)

    return con


def load_otel_runs(runs_dir: str | Path) -> Dict[str, duckdb.DuckDBPyConnection]:
    """
    Load OpenTelemetry data from multiple run directories.

    Args:
        runs_dir: Path to directory containing subdirectories with OTEL data
                  Each subdirectory should contain trace.jsonl, logs.jsonl, metrics.jsonl

    Returns:
        Dictionary mapping run name (subdirectory name) to DuckDB connection

    Example:
        >>> connections = load_otel_runs('/Users/arc/iavl-bench-data/sims')
        >>> # Returns: {'iavlx': <connection>, 'iavl1': <connection>}
        >>> connections['iavlx'].sql("SELECT count(*) FROM spans").show()
    """
    runs_dir = Path(runs_dir)
    connections = {}

    for subdir in runs_dir.iterdir():
        if subdir.is_dir():
            # Check if it has trace.jsonl to confirm it's a valid run directory
            if (subdir / 'trace.jsonl').exists():
                connections[subdir.name] = load_otel_data(subdir)

    return connections