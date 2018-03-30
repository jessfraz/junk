package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket/pcap"
)

var (
	wg sync.WaitGroup
)

func main() {
	ticker := time.NewTicker(30 * time.Second)

	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			ticker.Stop()
			fmt.Printf("Received %s, exiting.\n", sig.String())
			os.Exit(0)
		}
	}()

	// Process network packets.
	// Find all devices.
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}
	// Iterate over the devices and watch for packets.
	for _, device := range devices {
		wg.Add(1)
		go watchDevice(device)
	}

	// Get information from the /proc filesystem for the processes.
	go func(t *time.Ticker) {
		for range t.C {
			getProcInfo()
		}
	}(ticker)

	wg.Wait()
}
