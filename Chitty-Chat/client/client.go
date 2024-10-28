package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "Dis_Zyzz/Chitty-Chat/proto" // Import the generated protobuf package

	"google.golang.org/grpc"
)

func receiveMessages(client pb.ChittyChat_JoinChatClient) {
	for {
		msg, err := client.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			break
		}
		fmt.Printf("[%d] %s: %s\n", msg.Timestamp, msg.SenderId, msg.Text)
	}
}

func main() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewChittyChatClient(conn)

	fmt.Print("Enter your participant ID: ")
	reader := bufio.NewReader(os.Stdin)
	id, _ := reader.ReadString('\n')
	id = id[:len(id)-1] // remove newline

	joinStream, err := client.JoinChat(context.Background(), &pb.Participant{Id: id})
	if err != nil {
		log.Fatalf("Could not join chat: %v", err)
	}

	go receiveMessages(joinStream)

	for {
		time.Sleep(time.Millisecond * 500)
		fmt.Print("Enter message: ")
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]

		if len(text) > 128 {
			fmt.Println("Message too long! Maximum 128 characters.")
			continue
		}

		_, err := client.PublishMessage(context.Background(), &pb.PublishRequest{SenderId: id, Text: text})
		if err != nil {
			log.Printf("Could not send message: %v", err)
		}
	}
}
