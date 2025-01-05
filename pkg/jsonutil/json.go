package jsonutil

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func MustMarshal(v proto.Message) string {
	marshaller := protojson.MarshalOptions{EmitDefaultValues: true}
	bs, _ := marshaller.Marshal(v)
	return string(bs)
}
