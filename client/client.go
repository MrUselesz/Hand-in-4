package main

import (
	proto "Consensus/grpc"
	"log"
	"os"
	

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)



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

	conn, err := grpc.NewClient("localhost:6060", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}
	//when we know what we wwant to send recieve we can use this. 
	//client := proto.NewConsensusClient(conn)
	

}

func openLogFile(path string) (*os.File, error) {
	logFile, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Println("Failed")
	}
	return logFile, nil
}
