package main

import (
	pb "Dis_Zyzz/Chitty-Chat/proto"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"google.golang.org/grpc"
)

func init() {

	// log to console and file
	f, err := os.OpenFile("my.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)

	log.SetOutput(wrt)
}

// LamportClock structure to manage logical timestamps
type LamportClock struct {
	mu        sync.Mutex
	timestamp int64
}

func (lc *LamportClock) Increment() int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.timestamp++
	return lc.timestamp
}

func (lc *LamportClock) Update(receivedTime int64) int64 {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.timestamp = max(lc.timestamp, receivedTime) + 1
	return lc.timestamp
}

func max(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// ChittyChatServer structure manages chat clients and messages
type ChittyChatServer struct {
	pb.UnimplementedChittyChatServer
	clock        LamportClock
	participants map[string]chan *pb.Message
	mu           sync.Mutex
}

// JoinChat handles new participants
func (s *ChittyChatServer) JoinChat(participant *pb.Participant, stream pb.ChittyChat_JoinChatServer) error {
	s.mu.Lock()
	clientChan := make(chan *pb.Message, 100)
	s.participants[participant.Id] = clientChan
	s.mu.Unlock()

	// Broadcast notification that a new participant has joined
	timestamp := s.clock.Increment()
	joinMsg := &pb.Message{
		SenderId:  "System",
		Text:      fmt.Sprintf("Participant %s joined Chitty-Chat at Lamport time %d", strings.ReplaceAll(participant.Id, "\r", ""), timestamp),
		Timestamp: timestamp,
	}
	log.Print(joinMsg.Text)
	s.broadcastMessage(joinMsg)

	// Stream messages to this participant
	for msg := range clientChan {
		if err := stream.Send(msg); err != nil {
			break
		}
	}

	// Remove participant and broadcast notification that participant has left
	s.mu.Lock()
	delete(s.participants, participant.Id)
	s.mu.Unlock()

	timestamp = s.clock.Increment()

	return nil
}
func (s *ChittyChatServer) LeaveChat(ctx context.Context, req *pb.Participant) (*pb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var id = req.Id
	delete(s.participants, req.Id)

	timestamp := s.clock.Increment()

	leaveMsg := &pb.Message{
		SenderId:  "System",
		Text:      fmt.Sprintf("Participant %s left Chitty-Chat at Lamport time %d", strings.ReplaceAll(id, "\r", ""), timestamp),
		Timestamp: timestamp,
	}
	log.Print(leaveMsg.Text)
	s.broadcastMessage(leaveMsg)
	return &pb.Empty{}, nil
}

// PublishMessage handles incoming messages from clients
func (s *ChittyChatServer) PublishMessage(ctx context.Context, req *pb.PublishRequest) (*pb.Empty, error) {
	timestamp := s.clock.Increment()
	message := &pb.Message{
		SenderId:  req.SenderId,
		Text:      req.Text,
		Timestamp: timestamp,
	}
	log.Printf("[%d] [%s] %s\n", message.Timestamp, strings.ReplaceAll(message.SenderId, "\r", ""), strings.ReplaceAll(message.Text, "\r", ""))

	s.broadcastMessage(message)
	return &pb.Empty{}, nil
}

// broadcastMessage sends a message to all participants
func (s *ChittyChatServer) broadcastMessage(message *pb.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, ch := range s.participants {
		ch <- message
	}
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	chatServer := &ChittyChatServer{
		clock:        LamportClock{},
		participants: make(map[string]chan *pb.Message),
	}
	pb.RegisterChittyChatServer(grpcServer, chatServer)

	log.Println("Chitty-Chat server is running on port 50051 ...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
