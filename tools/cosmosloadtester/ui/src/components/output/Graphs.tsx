import React from 'react';
import Grid from '@mui/material/Grid';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Box from '@mui/material/Box';

import { ColorMap, scaleType } from 'src/types';
import { COLORS, /** PERCENTILE_KEYS */ } from 'src/constants';
import LineChart, { LinePoint } from 'src/components/charts/Lines';
import { useDivSize } from 'src/components/hooks/useDivSize';

import { LoadTestOutput } from "./data";

interface GraphProps {
    data: LoadTestOutput;
}

class Point {
    constructor(private x: number, private y: number, private dim: string, private xdata?: {[key: string]: number }) {}
    getX() { return this.x; }
    getY() { return this.y; }
    getDimension() { return this.dim; }
    getXAssociatedData() { return this.xdata; }
};

export default function Graphs(props: GraphProps) {
    const { data } = props;

    const timeToQPSRef = React.createRef<HTMLDivElement>();
    const timeToBytesRef = React.createRef<HTMLDivElement>();

    const { width, height } = useDivSize(timeToQPSRef); // define a shared and equal width/height for all graphs drawn

    const [timeToQPSData] = React.useState<LinePoint[]>(() => {
        // compute the data from 'data' to return a type the implements the
        // LinePoint interface maintaining a single dimension required for the
        // single line chart needed for time/QPS graph
        const t: LinePoint[] = [];
        data.forEach((d) => {
            d.per_sec.forEach((p) => {
                t.push(new Point(p.sec, p.qps, 'qps'));
            });
        });
        return t;
    });

    const [timeToBytesData] = React.useState<LinePoint[]>(() => {
        // compute the data from 'data' to return a type that implemnts the 
        // LinePoint interface while containing the different dimensions required
        // for a multiline graph.
        const t: LinePoint[] = [];
        data.forEach((d) => {
            d.per_sec.forEach((p) => {
                const ranking = p.bytes_rankings;
                t.push(new Point(p.sec, p.bytes, 'bytes', {
                    'p50 (avg)': ranking.p50.size,
                    'p75': ranking.p50.size,
                    'p90': ranking.p50.size,
                    'p95': ranking.p50.size,
                    'p99': ranking.p50.size,
                }));
            });
        });
        return t;
    });

    const [timeToBytesDimension, onHighlightBTimeToytesDimension] = React.useState<null | string>(null);

    const [dimensionToColorMap] = React.useState<ColorMap>({
        qps: COLORS.QPS,
        p50: COLORS.p50,
        p75: COLORS.p75,
        p90: COLORS.p90,
        p95: COLORS.p95,
        p99: COLORS.p99,
        bytes: COLORS.bytes,
    });

    const [yScale] = React.useState<scaleType>('linear');

    const onHighlightName = React.useCallback((name: string) => {
        if (name === timeToBytesDimension) onHighlightBTimeToytesDimension(null);
        else onHighlightBTimeToytesDimension(name);
    }, [timeToBytesDimension]);

    return (
        <Box sx={{ mt: 2 }}>
            <Grid container spacing={2}>
                <Grid item xs={12} sm={6}>
                    <Paper elevation={1} sx={{ textAlign: 'center', p: 2 }}>
                        <Typography variant="overline">
                            Transactions (QPS)
                        </Typography>
                        <Paper elevation={0}>
                            <Paper component="div" ref={timeToQPSRef} elevation={0}>
                                <LineChart
                                    target={timeToQPSRef}
                                    data={timeToQPSData}
                                    dimensionToColorMap={dimensionToColorMap}
                                    width={width}
                                    height={height}
                                    nameToHighlight={null}
                                    onHighlightName={() => { }}
                                    yScale={yScale}
                                    xAxisShortTitle={'s'}
                                    xAxisTitle={'Time (Seconds)'}
                                    yAxisShortTitle={'Tx'}
                                    yAxisTitle={'↑ Txs'}
                                />
                            </Paper>
                        </Paper>
                    </Paper>
                </Grid>
                <Grid item xs={12} sm={6}>
                    <Paper elevation={1} sx={{ textAlign: 'center', p: 2 }}>
                        <Typography variant="overline">
                            Bytes Sent
                        </Typography>
                        <Paper elevation={0}>
                            <Paper component="div" ref={timeToBytesRef} elevation={0}>
                                <LineChart
                                    target={timeToBytesRef}
                                    data={timeToBytesData}
                                    dimensionToColorMap={dimensionToColorMap}
                                    width={width}
                                    height={height}
                                    nameToHighlight={timeToBytesDimension}
                                    onHighlightName={onHighlightName}
                                    yScale={yScale}
                                    xAxisShortTitle={'t'}
                                    xAxisTitle={'Time (seconds)'}
                                    yAxisShortTitle={'B'}
                                    yAxisTitle={'↑ Bytes'}
                                />
                            </Paper>
                        </Paper>
                    </Paper>
                </Grid>
            </Grid>
        </Box>
    );
}
