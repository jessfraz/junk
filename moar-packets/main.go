package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
	// Find all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		// Print the device information.
		fmt.Printf("Name: %s\n", device.Name)
		fmt.Printf("Description: %s\n", device.Description)
		fmt.Printf("Devices addresses: %s\n", device.Description)
		for _, address := range device.Addresses {
			fmt.Printf("- IP address: %s\n", address.IP)
			fmt.Printf("- Subnet mask: %s\n", address.Netmask)
		}

		// Open the device.
		fmt.Println("Opening device...")
		handle, err := pcap.OpenLive(device.Name, 1024, false, 30*time.Second)
		if err != nil {
			log.Fatal(err)
		}
		defer handle.Close()

		// Use the handle as a packet source and process all packets.
		fmt.Println("Reading packets...")
		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			// Process the packet.
			fmt.Printf("packet: %#v\n", packet)
		}
	}
}
