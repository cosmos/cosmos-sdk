from concurrent import futures
import sys
import time

import grpc

import service_pb2
import service_pb2_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc

class ABCIListenerServiceServicer(service_pb2_grpc.ABCIListenerServiceServicer):
    """Implementation of ABCListener service."""

    def ListenBeginBlock(self, request, context):
        println("block_height: %d, req: %b, res: %b" % (request.block_height, request.req, req.res))
        return service_pb2.Empty()

    def ListenEndBlock(self, request, context):
        println("block_height: %d, req: %b, res: %b" % (req.block_height, request.req, req.res))
        return service_pb2.Empty()

    def ListenDeliverTx(self, request, context):
        println("block_height: %d, req: %b, res: %b" % (req.block_height, request.req, req.res))
        return service_pb2.Empty()

    def ListenStoreKVPair(self, request, context):
        println("block_height: %d, store_kv_pair: %b" % (req.block_height, request.store_kv_pair))
        return service_pb2.Empty()

def serve():
    # We need to build a health service to work with streaming-go-streaming
    health = HealthServicer()
    health.set("streaming", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    service_pb2_grpc.add_ABCIListenerServiceServicer_to_server(ABCIListenerServiceServicer(), server)
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
