package serializer

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

// ProtobufToJSON marshals a protobuf to JSON
func ProtobufToJSON(message proto.Message) (string, error) {
	marshler := jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: true,
		Indent:       " ",
		OrigName:     true,
	}
	return marshler.MarshalToString(message)
}
