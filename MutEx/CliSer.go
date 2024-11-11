package main

import (
	mutex "Dis_Zyzz/MutEx/mutex" // Import the generated package (adjust path based on your project structure)
	"context"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
)

func init() {

	// log to console and file
	f, err := os.OpenFile("mutex.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	wrt := io.MultiWriter(os.Stdout, f)

	log.SetOutput(wrt)
}

type server struct {
	mutex.UnimplementedMutexServiceServer
	nodeID       int
	lamportClock int64
	inCritical   bool
	replyCount   int
	peers        []mutex.MutexServiceClient
	mu           sync.Mutex
}

func (s *server) RequestAccess(ctx context.Context, req *mutex.RequestMessage) (*mutex.ReplyMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Node %d received request from Node %d at timestamp %d\n", s.nodeID, req.NodeId, req.Timestamp)

	if !s.inCritical && (s.lamportClock < req.Timestamp || (s.lamportClock == req.Timestamp && int32(s.nodeID) < req.NodeId)) {
		s.lamportClock++
		return &mutex.ReplyMessage{Granted: true, Timestamp: s.lamportClock}, nil
	} else {
		s.lamportClock++
		return &mutex.ReplyMessage{Granted: false, Timestamp: s.lamportClock}, nil
	}
}

func (s *server) Release(ctx context.Context, req *mutex.ReleaseMessage) (*mutex.Empty, error) {
	s.mu.Lock()
	s.inCritical = false
	s.replyCount = 0
	s.lamportClock++
	log.Printf("Node %d releasing at timestamp %d\n", s.nodeID, req.Timestamp)
	s.mu.Unlock()

	// Notify peers and process any pending requests
	s.processPendingRequests()

	return &mutex.Empty{}, nil
}

func (s *server) EnterCriticalSection() {
	s.mu.Lock()
	s.lamportClock++
	reqTimestamp := s.lamportClock
	s.mu.Unlock()

	log.Printf("Node %d requesting CS at timestamp %d\n", s.nodeID, reqTimestamp)

	// Send RequestAccess to all peers
	for _, peer := range s.peers {
		go func(p mutex.MutexServiceClient) {
			resp, err := p.RequestAccess(context.Background(), &mutex.RequestMessage{NodeId: int32(s.nodeID), Timestamp: reqTimestamp})
			if err == nil && resp.Granted {
				s.mu.Lock()
				s.replyCount++
				s.mu.Unlock()
			}
		}(peer)
	}

	// Wait for responses from all nodes
	s.mu.Lock()
	for s.replyCount < len(s.peers) {
		s.mu.Unlock()
		time.Sleep(100 * time.Millisecond) // Simulate waiting for responses
		s.mu.Lock()
	}
	s.mu.Unlock()

	// Enter the critical section
	s.inCritical = true
	log.Printf("Node %d entering critical section\n", s.nodeID)
}

func (s *server) processPendingRequests() {
	// This function can process queued requests from nodes waiting for CS
	log.Printf("Processing pending requests\n")
}

func startServer(nodeID int, address string) *server {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	s := &server{
		nodeID: nodeID,
	}
	mutex.RegisterMutexServiceServer(grpcServer, s)
	log.Printf("Starting server for Node %d on %s\n", nodeID, address)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
	return s
}

func startClient(nodeID int, address string, peers []string, s *server) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Set up peers (other nodes in the system)
	for _, peerAddr := range peers {
		peerConn, err := grpc.Dial(peerAddr, grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("could not connect to peer %s: %v", peerAddr, err)
		}
		peerClient := mutex.NewMutexServiceClient(peerConn)
		s.peers = append(s.peers, peerClient)
	}

	// Simulate the node requesting CS
	log.Printf("Node %d requesting critical section\n", nodeID)
	s.EnterCriticalSection()
}

func discover() []string {
	var slice []string
	slice = append(slice, "5-50051")
	slice = append(slice, "localhost:50052")
	slice = append(slice, "localhost:50053")
	return slice
}

func main() {
	slice := discover()
	user := strings.Split(slice[0], "-")
	nodeID, _ := strconv.Atoi(user[0]) // For example, this is Node 1
	address := user[1]                 // Port for this node
	peers := slice[1:]                 // Other nodes' addresses
	if len(slice) == 1 {
		os.Truncate("/path/to/your/file/crop.csv", 0)
	}
	s := startServer(nodeID, address)      // Start server for the node
	startClient(nodeID, address, peers, s) // Start client functionality

	// Keep the server running
	select {}
}
