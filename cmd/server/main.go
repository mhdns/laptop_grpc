package main

import (
	"context"
	"flag"
	"fmt"
	"grpc_youtube_tutorial/pb"
	"grpc_youtube_tutorial/service"
	"log"
	"net"

	"google.golang.org/grpc"
)

func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	log.Printf("--> unary interceptor: %v", info.FullMethod)

	return handler(ctx, req)
}

func streamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	log.Printf("--> stream interceptor: %v", info.FullMethod)
	return handler(srv, ss)
}

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("start server on port %v", *port)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
	)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	li, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		log.Fatalf("unable to create listener: %v", err)
	}

	err = grpcServer.Serve(li)
	if err != nil {
		log.Fatalf("unable to start: %v", err)
	}
}
