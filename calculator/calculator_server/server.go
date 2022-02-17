package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"time"

	"../calculatorpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

type server struct{}

func (*server) Sum(ctx context.Context, req *calculatorpb.SumRequest) (*calculatorpb.SumResponse, error) {

	result := req.GetFirstNumber() + req.GetSecondNumber()

	res := &calculatorpb.SumResponse{
		SumResult: result,
	}
	return res, nil
}

func (*server) PND(req *calculatorpb.PNDRequest, stream calculatorpb.CalculatorService_PNDServer) error {

	k := req.GetNumber()

	n := int32(2)

	for {
		if k <= 1 {
			break
		}
		for {
			if k%n == 0 {
				k = k / n
				res := &calculatorpb.PNDResponse{
					Prime: n,
				}
				stream.Send(res)
				time.Sleep(time.Second)
			} else {
				n++
				break
			}
		}
	}
	return nil
}

func (*server) Avg(stream calculatorpb.CalculatorService_AvgServer) error {
	count := 0
	sum := int32(0)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			avg := float64(sum) / float64(count)
			res := &calculatorpb.AvgResponse{
				Average: avg,
			}
			return stream.SendAndClose(res)
		}
		if err != nil {
			log.Fatalf("Streaming error: %v", err)
		}
		sum += req.GetNumber()
		count++
	}
}

func (*server) FindMaximum(stream calculatorpb.CalculatorService_FindMaximumServer) error {
	max := int32(math.MinInt32)
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Streaming error: %v", err)
			break
		}
		num := req.GetNumber()
		if num > max {
			max = num
			res := &calculatorpb.FindMaximumResponse{
				Maximum: max,
			}
			stream.Send(res)
		}
	}
	return nil
}

func (*server) SquareRoot(ctx context.Context, req *calculatorpb.SquareRootRequest) (*calculatorpb.SquareRootResponse, error) {
	fmt.Println("Received SquareRoot RPC")

	num := req.GetNumber()

	if num < 0 {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Received a negative number %v", num),
		)
	}
	res := &calculatorpb.SquareRootResponse{
		NumberRoot: math.Sqrt(float64(num)),
	}

	return res, nil
}

func main() {
	fmt.Println("Calculator Server")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	tls := false
	if tls {
		certFile := "ssl/server.crt"
		keyFile := "ssl/server.pem"
		creds, sslErr := credentials.NewServerTLSFromFile(certFile, keyFile)
		if sslErr != nil {
			log.Fatalf("Failed loading certificates: %v", sslErr)
			return
		}
		opts = append(opts, grpc.Creds(creds))
	}

	s := grpc.NewServer(opts...)
	calculatorpb.RegisterCalculatorServiceServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
