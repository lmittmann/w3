package w3_test

import (
	"log"

	"github.com/lmittmann/w3"
)

func ExampleDial() {
	client, err := w3.Dial("https://cloudflare-eth.com")
	if err != nil {
		log.Fatalf("Failed to connect to RPC endpoint: %v", err)
	}
	defer client.Close()
}
