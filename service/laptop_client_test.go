package service_test

import (
	"bufio"
	"context"
	"fmt"
	"grpc_youtube_tutorial/pb"
	"grpc_youtube_tutorial/sample"
	"grpc_youtube_tutorial/service"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestClientRating(t *testing.T) {
	t.Parallel()

	laptopStore := service.NewInMemoryLaptopStore()
	ratingStore := service.NewInMemoryRatingStore()

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddr := startTestLaptopServer(t, laptopStore, nil, ratingStore)
	laptopClient := newTestLaptopClient(t, serverAddr)

	stream, err := laptopClient.RateLaptop(context.Background())
	require.NoError(t, err)

	scores := []float64{8, 7.5, 10}
	expectedAverageScore := []float64{8, 7.75, 8.5}

	n := len(scores)
	for i := 0; i < n; i++ {
		req := &pb.RateLaptopRequest{
			LaptopId: laptop.GetId(),
			Score:    scores[i],
		}
		err := stream.Send(req)
		require.NoError(t, err)
	}

	err = stream.CloseSend()
	require.NoError(t, err)

	for idx := 0; ; idx++ {
		res, err := stream.Recv()
		if err == io.EOF {
			require.Equal(t, n, idx)
			return
		}
		require.NoError(t, err)
		require.Equal(t, laptop.GetId(), res.GetLaptopId())
		require.Equal(t, uint32(idx+1), res.GetRatedCount())
		require.Equal(t, expectedAverageScore[idx], res.GetAverageScore())
	}

}

func TestClientUploadImage(t *testing.T) {
	t.Parallel()

	testImageFolder := "../tmp"

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore(testImageFolder)

	laptop := sample.NewLaptop()
	err := laptopStore.Save(laptop)
	require.NoError(t, err)

	serverAddr := startTestLaptopServer(t, laptopStore, imageStore, nil)
	laptopClient := newTestLaptopClient(t, serverAddr)

	imagePath := fmt.Sprintf("%v/laptop.jpg", testImageFolder)
	file, err := os.Open(imagePath)
	require.NoError(t, err)
	defer file.Close()

	stream, err := laptopClient.UploadImage(context.Background())
	require.NoError(t, err)

	req := &pb.UploadImageRequest{
		Data: &pb.UploadImageRequest_Info{
			Info: &pb.ImageInfo{
				LaptopId:  laptop.GetId(),
				ImageType: filepath.Ext(imagePath),
			},
		},
	}

	err = stream.Send(req)
	require.NoError(t, err)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)
	size := 0
	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		req := &pb.UploadImageRequest{
			Data: &pb.UploadImageRequest_Chunk{
				Chunk: buffer[:n],
			},
		}

		size += n
		err = stream.Send(req)
		require.NoError(t, err)
	}

	res, err := stream.CloseAndRecv()
	require.NoError(t, err)
	require.NotZero(t, res.GetId())
	fSize, _ := strconv.Atoi(res.GetSize())
	require.Equal(t, size, fSize)

	savedImagePath := fmt.Sprintf("%s/%s%s", testImageFolder, res.GetId(), filepath.Ext(imagePath))
	require.FileExists(t, savedImagePath)
	require.NoError(t, os.Remove(savedImagePath))
}

func TestClientCreateLaptop(t *testing.T) {
	t.Parallel()

	laptopStore := service.NewInMemoryLaptopStore()
	serverAddr := startTestLaptopServer(t, laptopStore, nil, nil)
	laptopClient := newTestLaptopClient(t, serverAddr)

	laptop := sample.NewLaptop()
	expectedID := laptop.Id
	req := &pb.CreateLaptopRequest{
		Laptop: laptop,
	}

	res, err := laptopClient.CreateLaptop(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.Equal(t, expectedID, res.GetId())

	// Check if the laptop really stored in the store
	other, present := laptopStore.Find(expectedID)
	require.True(t, present)
	require.NotNil(t, other)

}

func startTestLaptopServer(t *testing.T, laptopStore service.LaptopStore, imageStore *service.DiskImageStore, ratingStore service.RatingStore) string {
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	grpcServer := grpc.NewServer()

	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	// :0 means any random available port
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	go grpcServer.Serve(listener)

	return listener.Addr().String()
}

func newTestLaptopClient(t *testing.T, serverAddrString string) pb.LaptopServiceClient {
	conn, err := grpc.Dial(serverAddrString, grpc.WithInsecure())
	require.NoError(t, err)

	return pb.NewLaptopServiceClient(conn)
}

func TestClientSearchCreateLaptop(t *testing.T) {
	t.Parallel()

	filter := &pb.Filter{
		MaxPriceUsd: 2000,
		MinCpuCores: 4,
		MinCpuGhz:   2.2,
		MinMemory:   &pb.Memory{Value: 8, Unit: pb.Memory_GIGABYTE},
	}

	store := service.NewInMemoryLaptopStore()
	expectedIds := make(map[string]bool)

	for i := 0; i < 6; i++ {
		laptop := sample.NewLaptop()

		switch i {
		case 0:
			laptop.PriceUsd = 2500
		case 1:
			laptop.Cpu.NumberCores = 2
		case 2:
			laptop.Cpu.MinGhz = 2.0
		case 3:
			laptop.Ram = &pb.Memory{Value: 4096, Unit: pb.Memory_MEGABYTE}
		case 4:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 4
			laptop.Cpu.MinGhz = 2.5
			laptop.Cpu.MaxGhz = 4.5
			laptop.Ram = &pb.Memory{Value: 16, Unit: pb.Memory_GIGABYTE}
			expectedIds[laptop.Id] = true
		case 5:
			laptop.PriceUsd = 1999
			laptop.Cpu.NumberCores = 6
			laptop.Cpu.MinGhz = 2.8
			laptop.Cpu.MaxGhz = 5.0
			laptop.Ram = &pb.Memory{Value: 64, Unit: pb.Memory_GIGABYTE}
			expectedIds[laptop.Id] = true
		}

		err := store.Save(laptop)
		require.NoError(t, err)
	}

	serverAddr := startTestLaptopServer(t, store, nil, nil)
	laptopClient := newTestLaptopClient(t, serverAddr)

	req := &pb.SearchLaptopRequest{Filter: filter}
	stream, err := laptopClient.SearchLaptop(context.Background(), req)
	require.NoError(t, err)

	found := 0
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}

		require.NoError(t, err)
		require.Contains(t, expectedIds, res.GetLaptop().GetId())

		found += 1
	}

	require.Equal(t, len(expectedIds), found)
}
