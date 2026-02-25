package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joshp123/pi-golang"
)

func main() {
	opts := pi.DefaultOneShotOptions()
	opts.Mode = pi.ModeSmart
	client, err := pi.StartOneShot(opts)
	if err != nil {
		log.Fatalf("start: %v", err)
	}
	defer client.Close()

	result, err := client.Run(context.Background(), pi.PromptRequest{Message: "Say hello in one sentence."})
	if err != nil {
		log.Fatalf("run: %v", err)
	}

	fmt.Println(result.Text)
}
