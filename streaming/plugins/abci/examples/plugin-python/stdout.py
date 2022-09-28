from concurrent import futures
import sys
import time

import grpc

import listener_pb2
import listener_pb2_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc

from pathlib import Path

class ABCIListenerServiceServicer(listener_pb2_grpc.ABCIListenerServiceServicer):
    """Implementation of ABCListener service."""
    home = str(Path.home())

    def ListenBeginBlock(self, request, context):
        filename = "{}/{}".format(home, 'abci_begin_block.txt')
        line = "{}:::{}:::{}\n".format(request.block_height, request.req, request.res)
        with open(filename, 'a') as f:
            f.write(line)

        return listener_pb2.Empty()

    def ListenEndBlock(self, request, context):
        filename = "{}/{}".format(home, 'abci_end_block.txt')
        line = "{}:::{}:::{}\n".format(request.block_height, request.req, request.res)
        with open(filename, 'a') as f:
            f.write(line)

        return listener_pb2.Empty()

    def ListenDeliverTx(self, request, context):
        filename = "{}/{}".format(home, 'abci_deliver_tx.txt')
        line = "{}:::{}:::{}\n".format(request.block_height, request.req, request.res)
        with open(filename, 'a') as f:
            f.write(line)

        return listener_pb2.Empty()

    def ListenStoreKVPair(self, request, context):
        filename = "{}/{}".format(home, 'abci_store_kv_pair.txt')
        line = "{}:::{}\n".format(request.block_height, request.store_kv_pair)
        with open(filename, 'a') as f:
            f.write(line)

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
