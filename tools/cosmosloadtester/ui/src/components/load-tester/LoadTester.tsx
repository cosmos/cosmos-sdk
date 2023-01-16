import React from 'react';
import Paper from '@mui/material/Paper';
import Typography from '@mui/material/Typography';
import Alert from '@mui/material/Alert';
import Button from '@mui/material/Button';
import * as timestamp_pb from 'google-protobuf/google/protobuf/timestamp_pb';


import { LoadtestServiceClient } from 'src/gen/orijtech/cosmosloadtester/v1/Loadtest_serviceServiceClientPb';
import { RunLoadtestRequest, RunLoadtestResponse } from 'src/gen/orijtech/cosmosloadtester/v1/loadtest_service_pb';
import { Spinner } from 'src/components/Spinner';
import Inputs from 'src/components/inputs/Inputs';
import Outputs from 'src/components/output/Outputs';

import { fields } from './Fields';

const service = new LoadtestServiceClient('');

export default function LoadTester() {
    const [running, setRunning] = React.useState(false);
    const [error, setError] = React.useState('');
    const [data, setData] = React.useState<RunLoadtestResponse.AsObject>();

    const submitRef = React.useRef<HTMLButtonElement>(null);

    const onFormSubmit = async (data: any) => {
        try {
            // reset previous response & error after new form submission
            setData(undefined);
            setError('');
            
            setRunning(true);
            const endpoints: string[] = data.endpoints?.split(',') || [];
            const request = new RunLoadtestRequest();
            request
                .setBroadcastTxMethod(data.broadcastTxMethod)
                .setClientFactory(data.clientFactory)
                .setDuration(new timestamp_pb.Timestamp().setSeconds(data.duration))
                .setEndpointSelectMethod(data.endpointSelectMethod)
                .setEndpointsList(endpoints)
                .setExpectPeersCount(data.expectPeersCount)
                .setMaxEndpointCount(data.maxEndpointCount)
                .setMinPeerConnectivityCount(data.minPeerConnectivityCount)
                .setPeerConnectTimeout(new timestamp_pb.Timestamp().setSeconds(data.peerConnectTimeout))
                .setSendPeriod(new timestamp_pb.Timestamp().setSeconds(data.sendPeriod))
                .setStatsOutputFilePath(data.statsOutputFilePath)
                .setTransactionCount(data.transactionCount)
                .setTransactionSizeBytes(data.transactionSizeBytes)
                .setTransactionsPerSecond(data.transactionsPerSecond)
                .setConnectionCount(data.connectionCount);

            // TODO: remove the next line that mocks a real fetch when testing is complete
            await (new Promise((resolve) => setTimeout(resolve, 3000)));
            setData({} as RunLoadtestResponse.AsObject);
            
            const result = await service.runLoadtest(request, null);

            setData(result.toObject());
            console.log(result.toObject());
        } catch (e: any) {
            console.log(e);
            setError(e.message);
        } finally {
            setRunning(false);
        }
    };

    return (
        <>
            <Spinner loading={running} />
            <Paper elevation={0} sx={{ mt: 3, p: 3 }}>
                <Alert severity='info' sx={{ mb: 3 }} variant='standard'>
                    <Typography variant="caption">
                        Enter TM Parameters
                    </Typography>
                </Alert>
                <Inputs
                    handleFormSubmission={onFormSubmit}
                    fields={fields}
                    submitRef={submitRef}
                />
                <Button
                    disableElevation={true}
                    disabled={running}
                    color={running ? 'inherit' : 'info'}
                    sx={{ textTransform: 'none' }}
                    variant='contained'
                    onClick={() => submitRef.current?.click()}
                >
                    {running ? 'Running load testing...' : 'Run load testing'}
                </Button>

                {/* display errors if any occured */}
                {error !== '' && <Typography component='h6' variant='caption'>{error}</Typography>}
                {/* render data if it exists */}
                {
                    data !== undefined &&
                    <>
                        <Alert severity='success' sx={{ my: 3 }}>
                            <Typography variant='caption'>
                                Load Testing Results
                            </Typography>
                        </Alert>
                        <Outputs data={data} />
                    </>
                }
            </Paper>
        </>
    );
}
