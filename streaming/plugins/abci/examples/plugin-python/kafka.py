from concurrent import futures
import sys
import time
import grpc
import socket

import listener_pb2
import listener_pb2_grpc


from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc

from pathlib import Path

from confluent_kafka import Producer

class ABCIListenerServiceServicer(listener_pb2_grpc.ABCIListenerServiceServicer):
    """Implementation of ABCListener service."""

    producer = Producer({'bootstrap.servers': "localhost:9092",
                         'client.id': socket.gethostname()})

    def ListenBeginBlock(self, request, context):
        self.producer.produce("raw_begin_block_req", key=str(request.block_height), value=str(request.req))
        self.producer.produce("raw_begin_block_res", key=str(request.block_height), value=str(request.res))
        return listener_pb2.Empty()

    def ListenEndBlock(self, request, context):
        self.producer.produce("raw_end_block_req", key=str(request.block_height), value=str(request.req))
        self.producer.produce("raw_end_block_res", key=str(request.block_height), value=str(request.res))
        return listener_pb2.Empty()

    def ListenDeliverTx(self, request, context):
        self. producer.produce("raw_deliver_tx_req", key=str(request.block_height), value=str(request.req))
        self.producer.produce("raw_deliver_tx_res", key=str(request.block_height), value=str(request.res))
        return listener_pb2.Empty()

    def ListenStoreKVPair(self, request, context):
        self.producer.produce("raw_state_change", key=str(request.block_height), value=str(request.store_kv_pair))
        return listener_pb2.Empty()

def serve():
    # We need to build a health service to work with streaming-go-streaming
    health = HealthServicer()
    health.set("streaming", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    listener_pb2_grpc.add_ABCIListenerServiceServicer_to_server(ABCIListenerServiceServicer(), server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port('127.0.0.1:1234')
    server.start()

    # Output information
    print("1|1|tcp|127.0.0.1:1234|grpc")
    sys.stdout.flush()

    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
