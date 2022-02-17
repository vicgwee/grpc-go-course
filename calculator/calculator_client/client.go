package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"../calculatorpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {

	fmt.Println("Calculator Client")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer cc.Close()

	c := calculatorpb.NewCalculatorServiceClient(cc)
	// fmt.Printf("Created client: %f", c)

	//doUnary(c)
	//doStreamingService(c)
	//doStreamingClient(c)

	doBidirectionalStreaming(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.SumRequest{
		FirstNumber:  1,
		SecondNumber: 20,
	}

	res, err := c.Sum(context.Background(), req)

	if err != nil {
		log.Fatalf("Calculator Request error:%v", err)
	}

	log.Printf("Calculator Sum Result: %v", res.SumResult)
}

func doStreamingService(c calculatorpb.CalculatorServiceClient) {
	req := &calculatorpb.PNDRequest{
		Number: 11247480,
	}

	stream, err := c.PND(context.Background(), req)

	if err != nil {
		log.Fatalf("PND Request error:%v", err)
	}

	for {
		msg, err := stream.Recv()

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Streaming error:%v", err)
		}
		log.Printf("Received:%v", msg.GetPrime())
	}
}

func doStreamingClient(c calculatorpb.CalculatorServiceClient) {
	stream, err := c.Avg(context.Background())

	if err != nil {
		log.Fatalf("Avg Request error: %v", err)
	}

	nums := []int32{101, 2, 3, 4, 5, 6}

	for _, num := range nums {
		req := &calculatorpb.AvgRequest{
			Number: int32(num),
		}
		err := stream.Send(req)
		if err != nil {
			log.Fatalf("Streaming error: %v", err)
		}
	}
	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Avg Close error: %v", err)
	}
	log.Printf("Avg result: %v", res.GetAverage())
}

func doBidirectionalStreaming(c calculatorpb.CalculatorServiceClient) {

	stream, err := c.FindMaximum(context.Background())

	if err != nil {
		log.Fatalf("FindMaximum Request error: %v", err)
	}

	nums := []int32{1, 2, 3, 5, 4, 6}
	var waitc chan struct{}

	// Send reqs
	go func() {
		for _, num := range nums {
			req := &calculatorpb.FindMaximumRequest{
				Number: int32(num),
			}
			err := stream.Send(req)
			if err != nil {
				log.Fatalf("Streaming error: %v", err)
			}
			time.Sleep(time.Second)
		}
		stream.CloseSend()
	}()

	// Receive numbers
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
			}
			if err != nil {
				log.Fatalf("Receive stream err: %v", err)
			}
			log.Printf("Received: %v", res.GetMaximum())
		}
	}()

	<-waitc
}

func doErrorUnary(c calculatorpb.CalculatorServiceClient) {
	num := int32(10)
	req := &calculatorpb.SquareRootRequest{
		Number: num,
	}

	res, err := c.SquareRoot(context.Background(), req)

	if err != nil {
		respErr, ok := status.FromError(err)
		if ok {
			fmt.Println(respErr.Message())
			fmt.Println(respErr.Code())
			if respErr.Code() == codes.InvalidArgument {
				fmt.Println("We sent a negative number!")
			}
		}
		log.Fatalf("Error calling Square Root:%v", err)
	}

	log.Printf("SquareRootResult: %v", res.GetNumberRoot())
}
