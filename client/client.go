package main

import (
	proto "Consensus/grpc"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// all Global values that we need for the node.
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

// this is the code that responds to talkToHost.
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

/*
* establishes connection and returns it
*
* @param address - address to which connection is to be established
* @returns a proto client
 */
func (s *ConsensusServer) connect(address string) (connection proto.ConsensusClient) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed connection on: " + address)
	}

	client := proto.NewConsensusClient(conn)

	return client
}

/*
* Made to run as a go rutine so it can run asynchronous
*
* @param address -the address to which it will call connect
* @param wg -a wait group to make sure it has permission from all nodes
 */

func (s *ConsensusServer) individualRequest(address string, wg *sync.WaitGroup) {
	defer wg.Done()

	connection := s.connect(address)

	response, err := connection.RequestPermission(context.Background(), &proto.Request{Id: s.id, Address: s.address})
	if err != nil {
		log.Fatalf("failed response on: "+address, err)
	}
	s.permissions = append(s.permissions, response.Id) //this might be dangerous ¯\_(ツ)_/¯
}

// request from all participents that they wish to write in the critical area.
func (s *ConsensusServer) reqeustAccess() { //this might need to be called as a goRutine to not stall other proccesses
	wg := new(sync.WaitGroup)

	for key, val := range s.nodeAddresses {
		if key == s.id {
			continue
		} //don't connect to itself
		wg.Add(1)
		go s.individualRequest(val, wg)
	}
	wg.Wait() //makes sure that all requests have returned and permission is thereby granted

	s.criticalArea()
}

func (s *ConsensusServer) RequestPermission(ctx context.Context, req *proto.Request) (*proto.Response, error) {
	if Contains(s.permissions, req.GetId()) {
		s.queue = append(s.queue, req.Id)
		//how do we make it wait...?
	}
	for {
		if (!Contains(s.permissions, req.GetId())){
			break
		}
	}
	return &proto.Response{Id: s.id}, nil
}

func (s *ConsensusServer) criticalArea() {
	fmt.Println(s.id, " has accessed the critical area")
	log.Println(s.id, " has accessed the critical area")
	time.Sleep(1 * time.Second)
	fmt.Println(s.id, " has left the critical area")

	//s.queue = slice[:0] //this might work?
}

// starts the server.
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

// This talks to the Host and that is hardcoded to be localhost 5000
func (s *ConsensusServer) talkToTheHost() {
	client := s.connect("localhost:5000")
	fmt.Println("We in!")
	stream, err := client.ToHost(context.Background(), &proto.Empty{})
	if err != nil {
		log.Fatalf("Error trying to connect to host", err)
	}
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error in recieving message ", err)
		}

		fmt.Println("We got the goods ", msg.GetId(), msg.GetNodeId(), msg.GetAddress())

		if s.id == 0 {
			s.id = msg.GetId()

		}
		s.nodeAddresses[msg.GetNodeId()] = msg.GetAddress()
	}

}

// this open the log
func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Log failed")
	}
	return logFile, nil
}

// simple util method for wheter a slice contains a
func Contains(s []int32, e int32) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
