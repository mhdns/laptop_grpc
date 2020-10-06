package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"grpc_youtube_tutorial/client"
	"grpc_youtube_tutorial/pb"
	"grpc_youtube_tutorial/sample"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createLaptop(laptopClient pb.LaptopServiceClient, laptop *pb.Laptop) {
	// laptop.Id = "493d03d3-9dcf-41e3-ab41-6ab17"
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := laptopClient.CreateLaptop(ctx, req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			log.Printf("laptop already exists: %v", err)
		} else {
			log.Fatalf("unable to create laptop: %v", err)
		}
		return
	}

	log.Printf("laptop created with id %v", res.GetId())
}

func searchLaptop(laptopClient pb.LaptopServiceClient, filter *pb.Filter) {
	log.Print("search filter: ", filter)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.SearchLaptopRequest{
		Filter: filter,
	}

	stream, err := laptopClient.SearchLaptop(ctx, req)
	if err != nil {
		log.Fatal("cannot search laptop: ", err)
	}

	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal("error while receiving stream: ", err)
		} else {
			log.Print("~ found: ", res.GetLaptop().GetName())
		}
	}

}

func testCreateLaptop(laptopClient pb.LaptopServiceClient) {
	createLaptop(laptopClient, sample.NewLaptop())
}

func testSearchLaptop(laptopClient pb.LaptopServiceClient) {
	for i := 0; i < 10; i++ {
		createLaptop(laptopClient, sample.NewLaptop())
	}

	searchLaptop(laptopClient, &pb.Filter{
		MaxPriceUsd: 5000,
	})
}

func uploadImage(laptopClient pb.LaptopServiceClient, laptopID string, imagePath string) {
	file, err := os.Open(imagePath)
	if err != nil {
		log.Fatal("cannot open image file: ", err)
	}
	defer file.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.UploadImage(ctx)
	if err != nil {
		log.Fatal("cannot upload image file: ", err)
	}

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptopID,
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	if err != nil {
		log.Fatal("cannot send image info: ", err)
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal("cannot read chunk to buffer: ", err, stream.RecvMsg(nil))
		}

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_Chunk{
				Chunk: buffer[:n],
			},
		}

		err = stream.Send(req)
		if err != nil {
			log.Fatal("cannot send chunk to server: ", err, stream.RecvMsg(nil))
		}
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatal("cannot recieve response: ", err)
	}

	log.Printf("image uploaded with id: %v, size: %v", res.GetId(), res.GetSize())
}

func testUploadImage(laptopClient pb.LaptopServiceClient) {
	laptop := sample.NewLaptop()
	createLaptop(laptopClient, laptop)
	uploadImage(laptopClient, laptop.GetId(), "tmp/laptop.jpg")
}

const (
	username        = "admin1"
	password        = "secret"
	refreshDuration = time.Minute
)

func authMethods() map[string]bool {
	const laptopServicePath = "/techschool.pcbook.LaptopService/"

	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

func main() {
	serverAddr := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddr)

	cc1, err := grpc.Dial(*serverAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}

	authClient := client.NewAuthClient(cc1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)
	if err != nil {
		log.Fatal("unable to create interceptor: ", err)
	}

	cc2, err := grpc.Dial(
		*serverAddr,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(interceptor.Unary()),
		grpc.WithStreamInterceptor(interceptor.Stream()),
	)
	if err != nil {
		log.Fatalf("unable to connect: %v", err)
	}

	laptopClient := pb.NewLaptopServiceClient(cc2)

	testSearchLaptop(laptopClient)

	testUploadImage(laptopClient)

	testRateLaptop(laptopClient)

}

func rateLaptop(laptopClient pb.LaptopServiceClient, laptopIDs []string, scores []float64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := laptopClient.RateLaptop(ctx)
	if err != nil {
		return fmt.Errorf("cannot rate laptop: %v", err)
	}

	waitResponse := make(chan error)
	// go routine to receive responses
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				log.Print("no more response")
				waitResponse <- nil
				return
			}
			if err != nil {
				waitResponse <- fmt.Errorf("cannot receive stream response: %v", err)
				return
			}
			log.Printf("recieved response: %v", res)
		}
	}()

	// send requests
	for i, laptopID := range laptopIDs {
		req := &pb.RateLaptopRequest{
			LaptopId: laptopID,
			Score:    scores[i],
		}

		err := stream.Send(req)
		if err != nil {
			return fmt.Errorf("cannot send stream request: %v - %v", err, stream.RecvMsg(nil))
		}

		log.Printf("sent request: %v", req)

	}

	err = stream.CloseSend()
	if err != nil {
		return fmt.Errorf("cannot close send stream: %v", err)
	}

	err = <-waitResponse
	return err
}

func testRateLaptop(laptopClient pb.LaptopServiceClient) {
	n := 3
	laptopIDs := make([]string, n)

	for i := 0; i < n; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.GetId()
		createLaptop(laptopClient, laptop)
	}

	scores := make([]float64, n)
	for {
		fmt.Print("rate laptop (y/n):")
		var answer string
		fmt.Scan(&answer)

		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}

		err := rateLaptop(laptopClient, laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}
	}
}
