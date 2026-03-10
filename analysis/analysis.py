"""Analysis functions for OpenTelemetry trace and metrics data."""

import duckdb
from datetime import datetime
from pydantic import BaseModel
import pandas as pd
import plotly.graph_objects as go
from typing import Dict

# Type alias for runs dictionary (run_name -> DuckDB connection)
Runs = Dict[str, duckdb.DuckDBPyConnection]


class BlockSummary(BaseModel):
    """Summary of block execution."""
    start_time: datetime
    end_time: datetime
    total_duration_seconds: float
    block_count: int


def block_summary(con: duckdb.DuckDBPyConnection) -> BlockSummary:
    """
    Generate a high-level summary of block execution.

    Args:
        con: DuckDB connection with spans view loaded

    Returns:
        BlockSummary object with:
        - start_time: Start time of first block
        - end_time: End time of last block
        - total_duration_seconds: Total duration from first block start to last block end
        - block_count: Total number of blocks processed

    Example:
        >>> from analysis.read_otel import load_otel_data
        >>> con = load_otel_data('/path/to/data')
        >>> summary = block_summary(con)
        >>> print(f"Processed {summary.block_count} blocks in {summary.total_duration_seconds:.2f}s")
    """
    result = con.sql("""
        SELECT
            MIN(start_time) AS start_time,
            MAX(end_time) AS end_time,
            EXTRACT(EPOCH FROM (MAX(end_time) - MIN(start_time))) AS total_duration_seconds,
            COUNT(*) AS block_count
        FROM spans
        WHERE span_name = 'Block' AND scope = 'cosmos-sdk/baseapp'
    """).fetchone()

    return BlockSummary(
        start_time=result[0],
        end_time=result[1],
        total_duration_seconds=result[2],
        block_count=result[3],
    )


def block_durations(con: duckdb.DuckDBPyConnection, span_name: str = 'Block') -> pd.DataFrame:
    """
    Get duration for each block.

    Args:
        con: DuckDB connection with spans view loaded

    Returns:
        DataFrame with columns:
        - block_number: Sequential block number (1-indexed)
        - duration_ms: Block duration in milliseconds

    Example:
        >>> from analysis.read_otel import load_otel_data
        >>> con = load_otel_data('/path/to/data')
        >>> df = block_durations(con)
        >>> df.head()
    """
    return con.sql("""
        SELECT
            ROW_NUMBER() OVER (ORDER BY start_time) AS block_number,
            EXTRACT(EPOCH FROM duration) * 1000 AS duration_ms
        FROM spans
        WHERE span_name = ? AND scope = 'cosmos-sdk/baseapp'
        ORDER BY start_time
    """, params=[span_name]).df()


def plot_block_durations(runs: Runs, span_name: str = 'Block') -> go.Figure:
    """
    Create a plotly line chart comparing block durations across runs.

    Args:
        runs: Dictionary mapping run name to DuckDB connection

    Returns:
        Plotly Figure object with block duration traces for each run

    Example:
        >>> from analysis.read_otel import load_otel_runs
        >>> runs = load_otel_runs('/path/to/data')
        >>> fig = plot_block_durations(runs)
        >>> fig.show()
    """
    fig = go.Figure()

    for run_name, con in runs.items():
        df = block_durations(con, span_name)
        fig.add_trace(go.Scatter(
            x=df['block_number'],
            y=df['duration_ms'],
            mode='lines',
            name=run_name,
            line=dict(width=2)
        ))

    fig.update_layout(
        title=f'{span_name} Duration Comparison',
        xaxis_title='Block Number',
        yaxis_title='Duration (ms)',
        hovermode='x unified',
        template='plotly_white'
    )

    return fig