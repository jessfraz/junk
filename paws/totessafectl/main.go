package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/jessfraz/junk/paws/moarpackets/types"
	"github.com/jessfraz/junk/paws/totessafe/reflector"
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
			continue
		}

		// Continue early if we have no data.
		if len(blob.Data) <= 0 {
			continue
		}

		// Try to unmarshal the blob and pretty print it.
		// Start with a guess it's a packet blob.
		if strings.HasPrefix(blob.Data, `{"type":"`) {
			var pb types.PacketBlob
			if err := json.Unmarshal([]byte(blob.Data), &pb); err != nil {
				log.Printf("parsing packet blob failed: %v", err)
				continue
			}

			// We have no error so pretty print the packet if it's an HTTP packet.
			if pb.FoundHTTP {
				pb.Layers = nil
				fmt.Println(pb)
			}
			continue
		}

		var pb types.ProcBlob
		if err := json.Unmarshal([]byte(blob.Data), &pb); err != nil {
			log.Printf("parsing proc blob failed: %v", err)
			continue
		}

		// We have no error so pretty print the proc blob.
		fmt.Println(pb)
		continue
	}
}
