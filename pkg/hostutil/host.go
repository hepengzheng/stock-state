package hostutil

import (
	"log"
	"os"
	"sync"
)

const port string = "9000"

var (
	address string
	once    sync.Once
)

func init() {
	once.Do(func() {
		host, err := os.Hostname()
		if err != nil {
			log.Fatalf("Failed to get hostname: %v", err)
		}
		address = host + ":" + port
	})
}

func GetLocalAddr() string {
	return address
}
