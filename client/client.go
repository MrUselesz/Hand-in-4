package main

import (
	proto "Consensus/grpc"
	"context"
	"fmt"
	"io"
	"time"

	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

//all Global values that we need for the node.
type ConsensusServer struct {
	proto.UnimplementedConsensusServer

	id            int32
	address       string
	queue         []int32
	permissions   []int32
	nodeAddresses map[int32]string
	nodeCount     int32
	connections   map[int32]proto.ConsensusClient
}


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

	server := &ConsensusServer{connections: make(map[int32]proto.ConsensusClient), nodeAddresses: make(map[int32]string)}

	server.start_server()

}
//this is the code that responds to talkToHost.
func (s *ConsensusServer) ToHost(req *proto.Empty, stream proto.Consensus_ToHostServer) error {
	fmt.Println("we recieved yout call")
	var temp [3]int32
	temp[0] = 4
	temp[1] = 2
	temp[2] = 3
	for i := 0; i < 3; i++ {
		select {
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Stream has ended")
		default:
			fmt.Println("Sendinge")
			err := stream.Send(&proto.NodeId{Id: temp[i]})
			if err != nil {
				return nil
			}
			fmt.Println("message sent!")
		}
	}

	return nil

}
//this when logic is added need a address input i have not added that cause i have just tested sending and recieving data
func (s *ConsensusServer) connect() (connection proto.ConsensusClient) {
	conn, err := grpc.NewClient("localhost:5000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	client := proto.NewConsensusClient(conn)

	return client
}

//request from all participents that they wish to write in the critical area.
func (s *ConsensusServer) reqeustAccess(){
	
	connection := s.connect()
	
	response, err := connection.RequestPermission(context.Background(), &proto.Request{Id: s.id, Address: s.address})
	if err != nil {
		log.Fatalf("Fuck ", err)
	}

	fmt.Println(response.GetId())
	s.criticalArea()
}

func(s *ConsensusServer) RequestPermission(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	return &proto.Response{Id: s.id}, nil
}

func (s *ConsensusServer) criticalArea(){
	fmt.Println(s.id, " has accessed the critical area")
	time.Sleep(1*time.Second)
	fmt.Println(s.id, " has left the critical area")
}


//starts the server.
func (s *ConsensusServer) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatalf("Failed to listen on port: ", s.address, err)

	}
	fmt.Println("Server is active")
	proto.RegisterConsensusServer(grpcServer, s)
	go func() {
		err = grpcServer.Serve(listener)

		if err != nil {
			log.Fatalf("Did not work")
		}
	}()

	if s.address != ":5000" {
		s.talkToTheHost()
		s.reqeustAccess()
	}
}
//This talks to the Host and that is hardcoded to be localhost 5000
func (s *ConsensusServer) talkToTheHost() {
	fmt.Println("To think is to be")

	client := s.connect()
	fmt.Println("We in!")
	stream, err := client.ToHost(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatalf("DUCK!", err)
	}
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("DUCKING! ", err)
		}

		fmt.Println("We got the goods ", msg.GetId(), msg.GetNodeId(), msg.GetAddress())

		if s.id == 0 {
			s.id = msg.GetId()

		}
		s.nodeAddresses[msg.GetNodeId()] = msg.GetAddress()
	}

}
//this open the log
func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Log failed")
	}
	return logFile, nil
}
