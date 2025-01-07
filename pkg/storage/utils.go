package storage

import (
	"fmt"
	"strconv"
)

func encodeValue(value int32) string {
	return fmt.Sprintf("%010d", value)
}

func decodeValue(s string) (int32, error) {
	atoi, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	itoa := strconv.Itoa(atoi)
	value, err := strconv.ParseInt(itoa, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(value), nil
}
