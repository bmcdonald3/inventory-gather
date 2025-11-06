// cmd/test/main.go (Temporary Test File)
package main

import (
	"log"
	"github.com/bmcdonald3/inventory-gather/pkg/inventory" // Use the actual path to your package
)

func main() {
	// The CollectAndPost function will run the simulated Redfish data
	// through the simulated (but sequential) API posting logic.
	log.Println("Starting API posting test...")

	// Hardcode a dummy IP since the Redfish part is simulated anyway.
	err := inventory.CollectAndPost("1.1.1.1") 
	if err != nil {
		log.Fatalf("Test failed: %v", err)
	}
	log.Println("Test finished. Check the API server logs for POST and PUT requests.")
}