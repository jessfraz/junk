package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jessfraz/paws/totessafe/reflector"
)

func main() {
	client := reflector.NewExternalReflectorClient("totessafe.contained.af", 14411)
	if err := client.Connect(); err != nil {
		log.Fatalf("connecting client to totessafe failed: %v", err)
	}
	for {
		blob, err := client.Client.Get(context.TODO(), &reflector.RequestType{})
		if err != nil {
			log.Fatalf("getting blob from totessafe client failed: %v", err)
			return
		}
		if len(blob.Data) > 0 {
			fmt.Println(blob.Data)
		}
	}
}
