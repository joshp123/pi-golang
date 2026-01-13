package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joshp123/pi-golang"
)

func main() {
	client, err := pi.Start(pi.DefaultOptions())
	if err != nil {
		log.Fatalf("start: %v", err)
	}
	defer client.Close()

	result, err := client.Run(context.Background(), "Say hello in one sentence.")
	if err != nil {
		log.Fatalf("run: %v", err)
	}

	fmt.Println(result.Text)
}
