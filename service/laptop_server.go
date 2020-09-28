package service

import (
	"bytes"
	"context"
	"errors"
	"grpc_youtube_tutorial/pb"
	"io"
	"log"
	"strconv"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LaptopServer is a server that provides laptop services
type LaptopServer struct {
	LaptopStore LaptopStore
	ImageStore  ImageStore
	RatingStore RatingStore
}

// NewLaptopServer returns pointer to a LaptopServer
func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{
		LaptopStore: laptopStore,
		ImageStore:  imageStore,
		RatingStore: ratingStore,
	}
}

// CreateLaptop method for Laptop Service
func (server *LaptopServer) CreateLaptop(ctx context.Context, req *pb.CreateLaptopRequest) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()

	log.Printf("recieved a create laptop request for laptop with id: %v", laptop.GetId())

	if len(laptop.GetId()) > 0 {
		_, err := uuid.Parse(laptop.GetId())
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop id is invalid: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Error(codes.Internal, "cannot generate new ID for laptop")
		}
		laptop.Id = id.String()
	}

	// Heavy processing
	// time.Sleep(6 * time.Second)

	// Don't save if request cancelled
	if ctx.Err() == context.Canceled {
		log.Print("request cancelled")
		return nil, status.Errorf(codes.Canceled, "request cancelled")
	}

	// Don't save if deadline exceeded
	if ctx.Err() == context.DeadlineExceeded {
		log.Print("deadline exceeded")
		return nil, status.Errorf(codes.DeadlineExceeded, "deadline exeeded")
	}

	err := server.LaptopStore.Save(laptop)
	if errors.Is(err, ErrAlreadyExists) {
		return nil, status.Errorf(codes.AlreadyExists, "unable to save data: %v", err)
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to save data: %v", err)
	}

	log.Printf("saved laptop with id: %v", laptop.GetId())

	return &pb.CreateLaptopResponse{
		Id: laptop.GetId(),
	}, nil
}

// SearchLaptop returns a laptop based on filter
func (server *LaptopServer) SearchLaptop(req *pb.SearchLaptopRequest, stream pb.LaptopService_SearchLaptopServer) error {
	filter := req.GetFilter()
	log.Printf("received a search laptop request with: %v", filter)

	err := server.LaptopStore.Search(filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{Laptop: laptop}

			err := stream.Send(res)
			if err != nil {
				return err
			}

			log.Printf("sent laptop with id: %s", laptop.GetId())
			return nil
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		log.Print(status.Error(codes.Canceled, "request is canceled"))
		return status.Error(codes.Canceled, "request is canceled")
	case context.DeadlineExceeded:
		log.Print(status.Error(codes.DeadlineExceeded, "request deadline exeeded"))
		return status.Error(codes.DeadlineExceeded, "request deadline exeeded")
	default:
		return nil
	}
}

// UploadImage service to upload laptop images
func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Print("cannot recieve image info: ", err)
		return status.Error(codes.Unknown, "cannot recieve image info")
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()

	log.Printf("recieved a upload image request with laptop ID %v and image type %v", laptopID, imageType)

	laptop, found := server.LaptopStore.Find(laptopID)
	if !found {
		return status.Error(codes.Internal, "cannot find laptop")
	}
	if laptop == nil {
		return status.Errorf(codes.InvalidArgument, "laptop %v doesn't exist", laptopID)
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		// Check error
		if err := contextError(stream.Context()); err != nil {
			return err
		}
		log.Print("waiting to recieve more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "cannot recieve chunk: %v", laptopID)
		}

		chunk := req.GetChunk()
		size := len(chunk)

		//  time.Sleep(time.Second)

		imageSize += size

		if imageSize > maxImageSize {
			return status.Error(codes.InvalidArgument, "image size too big: maximum 1mb")
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot write chunk to data: %v", err)
		}
	}

	imageID, err := server.ImageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot save image to the store: %v", err)
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: strconv.Itoa(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		return status.Errorf(codes.Internal, "unable to send response: %v", err)
	}

	log.Printf("image successfully saved with ID: %v and Size: %v", imageID, imageSize)
	return nil
}

// RateLaptop service allows user to rate laptops
func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}

		if err != nil {
			return status.Errorf(codes.Unknown, "unable to receive request stream: %v", err)
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("recieved a rate laptop request: laptop id %v, score: %v", laptopID, score)

		_, found := server.LaptopStore.Find(laptopID)
		if !found {
			return status.Errorf(codes.NotFound, "laptop with id, %v, not found", laptopID)
		}

		rating, err := server.RatingStore.Add(laptopID, score)
		if err != nil {
			return status.Errorf(codes.Internal, "unable to add rating to rating store: %v", err)
		}

		res := &pb.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			return status.Errorf(codes.Internal, "unable to send response: %v", err)
		}
	}
	return nil
}
