package main

import (
	proto "Consensus/grpc"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ConsensusServer struct {
	proto.UnimplementedConsensusServer
}

// global variables
var id int
var address string
var queue []int
var permissions []int
var nodeAddresses map[int]string
var nodeCount int
var connections map[int]proto.ConsensusClient

func main() {
	/* here if needed
	 	if runtime.GOOS == "windows" {
			newLine = "\r\n"
		} else {
			newLine = "\n"
		}
	*/
	file, err := openLogFile("../mylog.log")
	if err != nil {
		log.Fatalf("Not working")
	}
	log.SetOutput(file)

	server := &ConsensusServer{}

	server.start_server()
	fmt.Println("Decantes?")
	connect("localhost:5050", 5)
	fmt.Println("Work plz")
	recieve(connections[5])

}

func connect(newServerAddress string, newClientId int) {

	conn, err := grpc.NewClient(newServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	connections[newClientId] = proto.NewConsensusClient(conn)

}

func (s *ConsensusServer) start_server() {

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		//log.Fatalf("Failed to listen on port: ", address, err)

	}
	fmt.Println("Server is active")

	proto.RegisterConsensusServer(grpcServer, s)

	err = grpcServer.Serve(listener)

	if err != nil {
		log.Fatalf("Did not work")
	}

}

func recieve(test proto.ConsensusClient) {
	fmt.Println("To think is to be")
	req := &proto.Empty{}
	stream, err := test.ToHost(context.Background(), req)
	if err != nil {
		log.Fatalf("ahhhhhhhh")
	}
	for {
		fmt.Println("I am still alive I think?")
		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("ahhhhhhhh")
		}
		fmt.Println(message.GetId())
	}
	fmt.Println("I am still alive")
}

func (s *ConsensusServer) test(stream proto.Consensus_ToHostServer) error {
	fmt.Println("Server runs for real for real")
	if err := stream.Send(&proto.NodeId{Id: 5}); err != nil {
		return nil
	}
	return nil
}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Log failed")
	}
	return logFile, nil
}
