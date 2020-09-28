package serializer

import (
	"fmt"
	"io/ioutil"

	"github.com/golang/protobuf/proto"
)

// WriteProtobufToJSONFile writes a protobuf message to file
func WriteProtobufToJSONFile(message proto.Message, filename string) error {
	jsonString, err := ProtobufToJSON(message)
	if err != nil {
		return fmt.Errorf("unable to marshal message: %v", err)
	}

	err = ioutil.WriteFile(filename, []byte(jsonString), 0644)
	if err != nil {
		return fmt.Errorf("error occurred while writing file: %v", err)
	}

	return nil
}

// WriteProtobufToBinaryFile writes a protobuf message to file
func WriteProtobufToBinaryFile(message proto.Message, filename string) error {
	data, err := proto.Marshal(message)
	if err != nil {
		return fmt.Errorf("unable to marshal message: %v", err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("error occurred while writing file: %v", err)
	}

	return nil
}

// ReadProtobufFromBinaryFile reads a binary file containing a protobuf message
func ReadProtobufFromBinaryFile(filename string, message proto.Message) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error occurred while reading file: %v", err)
	}

	err = proto.Unmarshal(data, message)
	if err != nil {
		return fmt.Errorf("unable to unmarshal message: %v", err)
	}

	return nil
}
