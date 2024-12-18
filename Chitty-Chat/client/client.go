package main

import (
	pb "Dis_Zyzz/Chitty-Chat/proto" // Import the generated protobuf package
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"google.golang.org/grpc"
)

func receiveMessages(client pb.ChittyChat_JoinChatClient) {
	for {
		msg, err := client.Recv()
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			break
		}
		fmt.Printf("[%d] [%s] %s\n", msg.Timestamp, strings.ReplaceAll(msg.SenderId, "\r", ""), strings.ReplaceAll(msg.Text, "\r", ""))
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
	fmt.Print("Enter message (or /leave to exit): ")
	for {
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]

		if text == "/leave\r" {
			_, err := client.LeaveChat(context.Background(), &pb.Participant{Id: id})
			if err != nil {
				log.Printf("Could not leave chat: %v", err)
			}
			log.Printf("You have left the chat.")
			break
		}

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
