package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc/credentials"

	"../greetpb"

	"google.golang.org/grpc"
)

func main() {

	fmt.Println("Hello I'm a client")

	tls := false
	opts := grpc.WithInsecure()
	if tls {
		certFile := "ssl/ca.crt" // Certificate Authority Trust certificate
		creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
		if sslErr != nil {
			log.Fatalf("Error while loading CA trust certificate: %v", sslErr)
			return
		}
		opts = grpc.WithTransportCredentials(creds)
	}

	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)
	//fmt.Printf("Created client: %f", c)

	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	doBidirectionalStreaming(c)
}

func doUnary(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Victor",
			LastName:  "Gwee",
		},
	}
	res, err := c.Greet(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Victor",
			LastName:  "Gwee",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)

	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v", err)
	}
	for {
		msg, err := resStream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while streaming: %v", err)
		}
		log.Printf("Response from GreetManyTimes: %v", msg.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	names := []string{"a", "b", "c", "d"}

	requests := make([]*greetpb.LongGreetRequest, len(names))

	for i := 0; i < len(names); i++ {
		req := &greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: names[i],
				LastName:  names[len(names)-i-1],
			},
		}
		requests[i] = req
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet RPC: %v", err)
	}

	for _, req := range requests {
		stream.Send(req)
		time.Sleep(time.Second)
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while closing LongGreet RPC: %v", err)
	}
	fmt.Printf("LongGreet Response: %v", res.GetResult())
}

func doBidirectionalStreaming(c greetpb.GreetServiceClient) {
	names := []string{"a", "b", "c", "d"}

	requests := make([]*greetpb.GreetEveryoneRequest, len(names))

	for i := 0; i < len(names); i++ {
		req := &greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: names[i],
				LastName:  names[len(names)-i-1],
			},
		}
		requests[i] = req
	}

	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while calling GreetEveryone RPC: %v", err)
	}

	waitc := make(chan struct{})
	// send many messages to the server
	go func() {
		for _, req := range requests {
			fmt.Printf("Sent: %v \n", req)
			stream.Send(req)
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()

	// receive messages from the server
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while closing LongGreet RPC: %v", err)
				break
			}
			fmt.Printf("GreetEveryone Response: %v\n", res.GetResult())
		}
		close(waitc)
	}()

	<-waitc
}
