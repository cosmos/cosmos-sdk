from concurrent import futures
import sys
import time
import grpc
import snappy

from abci.v1 import grpc_pb2 as abci_v1
from abci.v1 import grpc_pb2_grpc as abci_v1_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc

from pathlib import Path


class ABCIListenerServiceServicer(abci_v1_grpc.ABCIListenerServiceServicer):
    """Implementation of ABCListener service."""

    out_dir = str(Path.home())

    def ListenBeginBlock(self, request, context):
        req_filename = "{}/{}.{}".format(self.out_dir, "begin-block-req", 'txt').lower()
        res_filename = "{}/{}.{}".format(self.out_dir, "begin-block-res", 'txt').lower()
        with open(req_filename, 'a') as f:
            line = "{}:::{}\n".format(request.req.header.height, request.req)
            f.write(line)
        with open(res_filename, 'a') as f:
            line = "{}:::{}\n".format(request.req.header.height, request.res)
            f.write(line)
        return abci_v1.Empty()

    def ListenEndBlock(self, request, context):
        req_filename = "{}/{}.{}".format(self.out_dir, "end-block-req", 'txt').lower()
        res_filename = "{}/{}.{}".format(self.out_dir, "end-block-res", 'txt').lower()
        with open(req_filename, 'a') as f:
            line = "{}:::{}\n".format(request.req.height, request.req)
            f.write(line)
        with open(res_filename, 'a') as f:
            line = "{}:::{}\n".format(request.req.height, request.res)
            f.write(line)
        return abci_v1.Empty()

    def ListenDeliverTx(self, request, context):
        req_filename = "{}/{}.{}".format(self.out_dir, "deliver-tx-req", 'txt').lower()
        res_filename = "{}/{}.{}".format(self.out_dir, "deliver-tx-res", 'txt').lower()
        with open(req_filename, 'a') as f:
            line = "{}:::{}\n".format(request.block_height, request.req)
            f.write(line)
        with open(res_filename, 'a') as f:
            line = "{}:::{}\n".format(request.block_height, request.res)
            f.write(line)
        return abci_v1.Empty()

    def ListenCommit(self, request, context):
        res_filename = "{}/{}.{}".format(self.out_dir, "commit-res", 'txt').lower()
        stc_filename = "{}/{}.{}".format(self.out_dir, "state-change", 'txt').lower()
        with open(res_filename, 'a') as f:
            line = "{}:::{}\n".format(request.block_height, request.res.data)
            f.write(line)
        with open(stc_filename, 'a') as f:
            line = "{}:::{}\n".format(request.block_height, request.change_set)
            f.write(line)
        return abci_v1.Empty()


def serve():
    # We need to build a health service to work with go-plugin
    health = HealthServicer()
    health.set("plugin", health_pb2.HealthCheckResponse.ServingStatus.Value('SERVING'))

    # Start the server.
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    abci_v1_grpc.add_ABCIListenerServiceServicer_to_server(ABCIListenerServiceServicer(), server)
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
