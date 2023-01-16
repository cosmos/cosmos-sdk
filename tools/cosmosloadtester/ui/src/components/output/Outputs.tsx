import { Typography } from '@mui/material';
import Box from '@mui/material/Box';
import Chip from '@mui/material/Chip';

import { RunLoadtestResponse as Result } from 'src/gen/orijtech/cosmosloadtester/v1/loadtest_service_pb';

import { sampleOutput } from './data';
import Graphs from './Graphs';

interface OutputProps {
    data: Result.AsObject;
};

interface ResultBarProps {
    title: string;
    value: string | number;
};

function ResultBar(props: ResultBarProps) {
    const { title, value } = props;
    return (
        <span style={{ marginRight: 8 }}>
            <Chip label={<Typography variant='caption' fontWeight='bold'>{title}</Typography>} variant='outlined' color='info'  sx={{ borderTopRightRadius: 0, borderBottomRightRadius: 0, borderRight: 'none' }} />
            <Chip label={value} variant='outlined' color='info'  sx={{ borderTopLeftRadius: 0, borderBottomLeftRadius: 0 }} />
        </span>
    );
}

export default function Outputs(props: OutputProps) {
    const { data } = props;
    return (
        <>
            <Box>
                <ResultBar title='avg bytes per second' value={data.avgBytesPerSecond} />
                <ResultBar title='avg tx per second' value={data.avgTxsPerSecond} />
                <ResultBar title='total bytes' value={data.totalBytes} />
                <ResultBar title='total time' value={`${data.totalTime?.seconds || 0} seconds`} />
                <ResultBar title='total tx' value={data.totalTxs} />
            </Box>
            <Graphs data={sampleOutput} />
        </>
    );
}
