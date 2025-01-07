package storage

import (
	"strconv"
)

func encodeValue(value int32) string {
	//bs := make([]byte, 4)
	//binary.LittleEndian.PutUint32(bs, uint32(value))
	//return string(bs)

	return strconv.FormatInt(int64(value), 10)
}

func decodeValue(s string) int32 {
	//return int32(binary.LittleEndian.Uint32([]byte(s)))
	result, _ := strconv.ParseInt(s, 0, 32)
	return int32(result)
}
