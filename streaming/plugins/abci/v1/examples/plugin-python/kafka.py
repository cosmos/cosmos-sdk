from concurrent import futures
import sys
import time
import grpc
import socket
import logging

from abci.v1 import grpc_pb2 as abci_v1
from abci.v1 import grpc_pb2_grpc as abci_v1_grpc

from grpc_health.v1.health import HealthServicer
from grpc_health.v1 import health_pb2, health_pb2_grpc

from confluent_kafka import Producer, KafkaError, KafkaException


class ABCIListenerServiceServicer(abci_v1_grpc.ABCIListenerServiceServicer):
    """Implementation of ABCListener service."""

    producer = Producer({'bootstrap.servers': "localhost:9092",
                         'client.id': socket.gethostname(),
                         'acks': "all",
                         'enable.idempotence': "true",
                         'max.in.flight.requests.per.connection': 1,
                         'linger.ms': 5,
                         'message.max.bytes': 20485760})

    # broker config to increase message max size
    # message.max.bytes: 204857600
    # socket.request.max.bytes: 204857600
    # replica.fetch.max.bytes: 20485760

    def acked(self, err, msg):
        """
        Reports the failure or success of a message delivery.
        Args:
            err (KafkaError): The error that occurred on None on success.
            msg (Message): The message that was produced or failed.
        """

        if err is not None:
            logging.error("Failed to deliver message: %s: %s" % (str(msg), str(err)))
            raise err
        else:
            logging.info("Message produced: %s" % (str(msg)))

    # Wait up to 1 second for events. Callbacks will be invoked during
    # this method call if the message is acknowledged.
    producer.poll(1)

    def ListenBeginBlock(self, request, context):
        block_height = request.req.header.height
        self.producer.produce("raw_begin_block_req", key=str(block_height), value=str(request.req), callback=self.acked)
        self.producer.produce("raw_begin_block_res", key=str(block_height), value=str(request.res), callback=self.acked)
        return abci_v1.Empty()

    def ListenEndBlock(self, request, context):
        block_height = request.req.height
        self.producer.produce("raw-end-block-req", key=str(block_height), value=str(request.req), callback=self.acked)
        self.producer.produce("raw-end-block-res", key=str(block_height), value=str(request.res), callback=self.acked)
        return abci_v1.Empty()

    def ListenDeliverTx(self, request, context):
        block_height = request.block_height
        self.producer.produce("raw-deliver-tx-req", key=str(block_height), value=str(request.req), callback=self.acked)
        self.producer.produce("raw-deliver-tx-res", key=str(block_height), value=str(request.res), callback=self.acked)
        return abci_v1.Empty()

    def ListenCommit(self, request, context):
        block_height = request.block_height
        self.producer.produce("raw-commit-res", key=str(block_height), value=str(request.res), callback=self.acked)
        self.producer.produce("raw-state-change", key=str(block_height), value=str(request.change_set),
                              callback=self.acked)
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
