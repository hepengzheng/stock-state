package hostutil

import (
	"fmt"
	"testing"
)

func TestGetLocalAddr(t *testing.T) {
	res := GetLocalAddr()
	fmt.Println(res)
}
