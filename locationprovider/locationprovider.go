package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"

	"github.com/hsmtkk/qiita-cloud-run-grpc-acl/proto"
	"google.golang.org/grpc"
)

type server struct {
	proto.UnimplementedLocationServiceServer
}

func (s *server) SayHello(ctx context.Context, in *proto.LocationRequest) (*proto.LocationResponse, error) {
	var longitude int32 = int32(rand.Intn(180)) - 90
	var latitude int32 = int32(rand.Intn(360)) - 180
	return &proto.LocationResponse{Longitude: longitude, Latitude: latitude}, nil
}

func main() {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("failed to parse %s as int; %v", portStr, err)
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to listen port %d; %v", port, err)
	}
	s := grpc.NewServer()
	proto.RegisterLocationServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve; %v", err)
	}
}
