package jsonutil

import (
	"google.golang.org/protobuf/proto"
)

func MustMarshal(v proto.Message) string {
	bs, _ := proto.Marshal(v)
	return string(bs)
}
