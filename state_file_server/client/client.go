package client

import (
	"context"
	"io"
	"log"

	pb "github.com/cosmos/cosmos-sdk/state_file_server/grpc/v1beta"

	"google.golang.org/grpc"
)

func NewClient(conn *grpc.ClientConn) (pb.StateFileClient, error) {
	client := pb.NewStateFileClient(conn)
}

func main() {
	// dial server
	conn, err := grpc.Dial(":50005", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("can not connect with server %v", err)
	}


	// create stream
	client := pb.NewStateFileClient(conn)

	in := &pb.StreamRequest{Id: 1}
	stream, err := client.(context.Background(), in)
	if err != nil {
		log.Fatalf("open stream error %v", err)
	}

	done := make(chan bool)

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}
			log.Printf("Resp received: %s", resp.Result)
		}
	}()

	<-done //we will wait until all response is received
	log.Printf("finished")
}