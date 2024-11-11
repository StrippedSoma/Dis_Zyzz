package main

import (
	mutex "Dis_Zyzz/MutEx/MutEx" // Import the generated package (adjust path based on your project structure)
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
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

type PeerClient struct {
	client  mutex.MutexServiceClient // gRPC client
	address string                   // Address of this peer
}

type server struct {
	mutex.UnimplementedMutexServiceServer
	nodeID       int
	lamportClock int64
	inCritical   bool
	replyCount   int
	peers        []PeerClient
	mu           sync.Mutex
	filepath     string
}

func (s *server) RequestAccess(ctx context.Context, req *mutex.RequestMessage) (*mutex.ReplyMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Node %d received request from Node %d at timestamp %d\n", s.nodeID, req.NodeId, req.Timestamp)

	if !s.inCritical && (s.lamportClock < req.Timestamp || (s.lamportClock == req.Timestamp && int32(s.nodeID) < req.NodeId)) {
		s.lamportClock++
		log.Printf("Node %d granted access at timestamp %d\n", s.nodeID, s.lamportClock)
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

func (s *server) Quit(ctx context.Context, req *mutex.QuitMessage) (*mutex.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Node %d quitting from port %s\n", req.NodeId, req.Port)

	// Update the file to mark the client's port as available again
	err := s.updatePortStatus(req.Port, false)
	if err != nil {
		log.Printf("Failed to update port status for quitting client: %v", err)
		return &mutex.Empty{}, err
	}

	// Remove the client from the peers list if needed (depends on how you're tracking peers)
	s.removePeer(req.Port)

	return &mutex.Empty{}, nil
}

func (s *server) removePeer(port string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	newPeers := []PeerClient{}
	for _, peer := range s.peers {
		if peer.address != port {
			newPeers = append(newPeers, peer)
		}
	}
	s.peers = newPeers
}

func (s *server) Shutdown() {
	if len(s.peers) == 0 {
		log.Println("No peers to notify on shutdown.")
		os.Remove("mutex.log")
		return
	}
	_, err := s.peers[0].client.Quit(context.Background(), &mutex.QuitMessage{
		NodeId: int32(s.nodeID),
		Port:   s.peers[0].address, // This should be the address or port string
	})
	if err != nil {
		log.Printf("Failed to notify server of quit: %v", err)
	}
}

func (s *server) EnterCriticalSection() {
	s.mu.Lock()
	s.lamportClock++
	reqTimestamp := s.lamportClock
	s.mu.Unlock()

	log.Printf("Node %d requesting CS at timestamp %d\n", s.nodeID, reqTimestamp)

	// Send RequestAccess to all peers
	for _, peer := range s.peers {
		go func(p PeerClient) {
			resp, err := p.client.RequestAccess(context.Background(), &mutex.RequestMessage{NodeId: int32(s.nodeID), Timestamp: reqTimestamp})
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
	s.Release(context.Background(), &mutex.ReleaseMessage{Timestamp: reqTimestamp})
}

func (s *server) processPendingRequests() {
	// This function can process queued requests from nodes waiting for CS
	log.Printf("Node %d now processing pending requests\n", s.nodeID)
}

func startServer(nodeID int, address string, filepath string) *server {
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	s := &server{
		nodeID:   nodeID,
		filepath: filepath,
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
		s.peers = append(s.peers, PeerClient{
			client:  mutex.NewMutexServiceClient(peerConn),
			address: peerAddr,
		})
	}

	// Simulate the node requesting CS
	log.Printf("Node %d requesting critical section\n", nodeID)

	s.EnterCriticalSection()
}

// Function to parse the file, increment user count, find an available port, and update the file
func handleClientRequest(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Scan()
	userCount, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Printf("failed to parse user count: %v", err)
		return nil, err
	}

	// Increment the user count
	userCount++

	// Parse the ports
	ports := make(map[string]bool) // Port number -> occupied
	occupiedPorts := []string{}    // List of occupied ports for listening

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Split(line, "-")
		if len(parts) != 2 {
			continue
		}

		port := parts[0]
		occupied := parts[1] == "y"
		ports[port] = occupied

		if occupied {
			occupiedPorts = append(occupiedPorts, port)
		}
	}

	// Find an available port
	availablePort := ""
	for port, occupied := range ports {
		if !occupied {
			availablePort = port
			ports[port] = true
			occupiedPorts = append(occupiedPorts, port)
			break
		}
	}

	if availablePort == "" {
		log.Printf("No available ports found")
	}

	// Prepare updated content to save back to the file
	var newFileContent strings.Builder
	newFileContent.WriteString(fmt.Sprintf("%d\n", userCount))
	for port, occupied := range ports {
		status := "n"
		if occupied {
			status = "y"
		}
		newFileContent.WriteString(fmt.Sprintf("%s-%s\n", port, status))
	}

	// Save the updated file content
	err = os.WriteFile(filePath, []byte(newFileContent.String()), 0644)
	if err != nil {
		log.Printf("failed to write to file: %v", err)
		return nil, err
	}
	occupiedPorts = append(occupiedPorts, strconv.Itoa(userCount))
	return occupiedPorts, nil
}

func (s *server) updatePortStatus(port string, occupied bool) error {
	file, err := os.Open(s.filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := []string{}
	scanner.Scan()

	lines = append(lines, scanner.Text())

	// Parse and update each line
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		parts := strings.Split(line, "-")
		if len(parts) != 2 {
			continue
		}

		currentPort := parts[0]
		status := "n"
		if parts[1] == "y" {
			status = "y"
		}

		if currentPort == port {
			if occupied {
				status = "y"
			} else {
				status = "n"
			}
		}

		lines = append(lines, fmt.Sprintf("%s-%s", currentPort, status))
	}

	// Save the updated file content
	return os.WriteFile(s.filepath, []byte(strings.Join(lines, "\n")), 0644)
}

func main() {
	filePath := "users.txt"
	occupiedPorts, err := handleClientRequest(filePath)
	if err != nil {
		log.Printf("Error, failed to handle client request: %v", err)
	}

	nodeID, _ := strconv.Atoi(occupiedPorts[len(occupiedPorts)-1])
	address := occupiedPorts[len(occupiedPorts)-2] // Port for this node
	peers := occupiedPorts[:len(occupiedPorts)-2]  // Other nodes' addresses
	if len(peers) == 0 {

	}
	s := startServer(nodeID, address, filePath) // Start server for the node
	startClient(nodeID, address, peers, s)      // Start client functionality

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Block and wait for signal
	go func() {
		<-sigChan
		log.Printf("Node %d received shutdown signal. Quitting...", s.nodeID)

		// Gracefully notify server that this client is quitting
		s.Shutdown()

		// Wait a moment to allow cleanup to complete
		time.Sleep(1 * time.Second)
		os.Exit(0)
	}()
	// Keep the server running
	select {}
}
