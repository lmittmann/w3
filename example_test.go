package w3_test

import (
	"fmt"
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

func ExampleI() {
	fmt.Printf("%v wei\n", w3.I("0x2b98d99b09e3c000"))
	fmt.Printf("%v wei\n", w3.I("3141500000000000000"))
	fmt.Printf("%v wei\n", w3.I("3.1415 ether"))
	fmt.Printf("%v wei\n", w3.I("31.415 gwei"))

	// Output:
	// 3141500000000000000 wei
	// 3141500000000000000 wei
	// 3141500000000000000 wei
	// 31415000000 wei
}
