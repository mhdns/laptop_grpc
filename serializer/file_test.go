package serializer_test

import (
	"grpc_youtube_tutorial/pb"
	"grpc_youtube_tutorial/sample"
	"grpc_youtube_tutorial/serializer"
	"testing"
)

func TestFileSerializer(t *testing.T) {
	t.Parallel()

	binaryFile := "../tmp/laptop.bin"
	jsonFile := "../tmp/laptop.json"

	laptop1 := sample.NewLaptop()
	err := serializer.WriteProtobufToBinaryFile(laptop1, binaryFile)
	if err != nil {
		t.Error(err)
	}

	err = serializer.WriteProtobufToJSONFile(laptop1, jsonFile)
	if err != nil {
		t.Error(err)
	}

	laptop2 := &pb.Laptop{}
	err = serializer.ReadProtobufFromBinaryFile(binaryFile, laptop2)
	if err != nil {
		t.Error(err)
	}

	// if laptop1 != laptop2 {
	// 	t.Error("\n", laptop1, "\n", laptop2)
	// }
}
