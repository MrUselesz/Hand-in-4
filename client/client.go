package main

import (
	proto "Consensus/grpc"
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"
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

var newLine string

// all Global values that we need for the node. Basically a node "class"
type ConsensusServer struct {
	proto.UnimplementedConsensusServer

	id               int32
	address          string
	queue            []int32
	permissions      []int32
	nodeAddresses    map[int32]string
	nodeCount        int32
	nodePort         int32
	connections      map[int32]proto.ConsensusClient
	isTryingCritical bool
}

func main() {

	//For trimming "     strings     "
	if runtime.GOOS == "windows" {
		newLine = "\r\n"
	} else {
		newLine = "\n"
	}

	//Connect the node to the log
	file, err := openLogFile("../mylog.log")
	if err != nil {
		log.Fatalf("Not working")
	}
	log.SetOutput(file)

	//Initialize node
	server := &ConsensusServer{connections: make(map[int32]proto.ConsensusClient), nodeAddresses: make(map[int32]string)}
	server.start_server()

}

// This is the code that responds to talkToHost.
func (s *ConsensusServer) ToHost(req *proto.Empty, stream proto.Consensus_ToHostServer) error {

	//Host keeps track of how many nodes joined
	s.nodeCount++

	//As well as the latest available port (host is 5000, each next node has port 5000 + (n-1) where n is the number of nodes )
	s.nodePort++

	//Case for first node, as the host currently does not keep track of any addresses (a node never maps its own address)
	if len(s.nodeAddresses) == 0 {
		err := stream.Send(&proto.NodeId{Id: s.nodeCount, NodeId: s.id, Address: fmt.Sprintf(":%v", s.nodePort), NodeAddress: s.address})
		if err != nil {
			return nil
		}

	} else { //Standard procedure, host assigning a node with an ID and port

		for nodeId, nodeAddress := range s.nodeAddresses {

			select {
			case <-stream.Context().Done():
				return status.Error(codes.Canceled, "Stream has ended")
			default:
				//Send message back to new node that requested it.
				//Id and Address are to be assigned to the new node. Technically only needs to be sent once but hey

				//NodeId and NodeAddress are the attributes of another node. This message is sent as many times as there are currently mapped nodes,
				//resulting in the new node being aware of all its peers
				err := stream.Send(&proto.NodeId{Id: s.nodeCount, NodeId: nodeId, Address: fmt.Sprintf(":%v", s.nodePort), NodeAddress: nodeAddress})
				if err != nil {
					return nil
				}
			}
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
	s.permissions = append(s.permissions, response.Id) //Register a received permission from a node with ID response.Id
}

// Request critical area permission from all participants.
func (s *ConsensusServer) reqeustAccess() {

	log.Println(s.id, " is requesting access to the critical area")

	s.isTryingCritical = true
	wg := new(sync.WaitGroup)

	for key, val := range s.nodeAddresses {
		if key == s.id {
			continue
		} //Don't connect to itself. Impossible but just in case :)
		wg.Add(1)
		go s.individualRequest(val, wg)
	}
	wg.Wait() //makes sure that all requests have returned and permission is thereby granted

	s.criticalArea()
}

func (s *ConsensusServer) RequestPermission(ctx context.Context, req *proto.Request) (*proto.Response, error) {

	//If the node requesting permission is not known to the one taking the request
	//it gets added to the recipient's list of addresses

	if _, ok := s.nodeAddresses[req.GetId()]; !ok {

		s.nodeAddresses[req.GetId()] = strings.Trim(req.GetAddress(), newLine)

	}

	//If the node that takes the request already got permission from the requesting node,
	//the requesting node gets put in the queue
	if Contains(s.permissions, req.GetId()) {
		s.queue = append(s.queue, req.GetId())

	} else if s.id < req.GetId() && s.isTryingCritical {

		//This else if statement checks for ID seniority. In case two nodes request permission
		//from each other simultaneously, the older node (lesser ID) will always get the permission first, while
		//the younger node gets put in its queue
		s.queue = append(s.queue, req.GetId())

	}

	//"Waiter" for loop. Basically, a node flushes its permissions and its queue
	//upon being done with the critical area, and gives permissions to nodes in its
	//queue. So, once the queue is flushed, this for loop no longer blocks the request
	for {
		if !Contains(s.queue, req.GetId()) {
			break
		}
	}

	//Give permission to whomsoever may have requested it
	log.Println(s.id, " is giving permission to: ", req.GetId())
	return &proto.Response{Id: s.id}, nil
}

func (s *ConsensusServer) criticalArea() {

	log.Println(s.id, " has accessed the critical area")

	//Simulate doing something critical to the area
	time.Sleep(5 * time.Second)

	log.Println(s.id, " has left the critical area")

	log.Println(s.queue, " is the current queue of: ", s.id)

	//Flush queue and permission lists
	s.queue = s.queue[:0]
	s.permissions = s.permissions[:0]

	//Not trying to access the critical area for a little bit
	s.isTryingCritical = false
}

// starts the server.
func (s *ConsensusServer) start_server() {

	//If parameter "host" is included, this node is the host
	if len(os.Args) > 1 && os.Args[1] == "host" {

		s.id = 1

		s.address = ":5000"

		//Counter needed for latest available port
		s.nodePort = 5000

		//Guess what this one means
		s.nodeCount = 1

	} else {

		//If not host, ask the host who you are
		s.talkToTheHost()

	}

	//Rest of the method is just initializing a node with the
	//hardcoded or host-provided information

	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		log.Fatalf("Failed to listen on port: ", s.address, err)

	}
	proto.RegisterConsensusServer(grpcServer, s)
	go func() {
		err = grpcServer.Serve(listener)

		if err != nil {
			log.Fatalf("Did not work")
		}
	}()

	//Runs perpetually until program is killed,
	//requests access almost immediately whenever it can

	for {
		s.reqeustAccess()
		time.Sleep(1 * time.Second)
	}
}

// This talks to the Host and that is hardcoded to be localhost 5000
func (s *ConsensusServer) talkToTheHost() {

	//
	client := s.connect("localhost:5000")
	s.nodeAddresses[1] = ":5000"
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

		//0 is default value for our Id, so this runs once when the node does not know itself,
		//then proceeds to set Id and port (address) values in its struct
		if s.id == 0 {

			s.id = msg.GetId()
			s.address = strings.Trim(msg.GetAddress(), "\n")

		}

		//Maps address of a peer node to the relevant Id once it got the info from host
		s.nodeAddresses[msg.GetNodeId()] = strings.Trim(msg.GetNodeAddress(), newLine)
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
