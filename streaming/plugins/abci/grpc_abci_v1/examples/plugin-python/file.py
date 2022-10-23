from concurrent import futures
import sys
import time

import grpc

import abci_listener_pb2
import abci_listener_pb2_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc

from pathlib import Path

class ABCIListenerServiceServicer(abci_listener_pb2_grpc.ABCIListenerServiceServicer):
    """Implementation of ABCListener service."""

    out_dir = str(Path.home())

    def Listen(self, request, context):
        filename = "{}/{}.{}".format(self.out_dir, request.event_type, 'txt').lower()
        line = "{}:::{}\n".format(request.block_height, request.data)
        with open(filename, 'a') as f:
            f.write(line)

        return abci_listener_pb2.Empty()

def serve():
    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set("plugin", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    abci_listener_pb2_grpc.add_ABCIListenerServiceServicer_to_server(ABCIListenerServiceServicer(), server)
    health_pb2_grpc.add_HealthServicer_to_server(health, server)
    server.add_insecure_port('127.0.0.1:1234')
    server.start()

    # Output handshake information
    # https://github.com/hashicorp/go-plugin/blob/master/docs/guide-plugin-write-non-go.md#4-output-handshake-information
    print("1|1|tcp|127.0.0.1:1234|grpc")
    sys.stdout.flush()

    try:
        while True:
            time.sleep(60 * 60 * 24)
    except KeyboardInterrupt:
        server.stop(0)

if __name__ == '__main__':
    serve()
