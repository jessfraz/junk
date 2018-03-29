package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	wg sync.WaitGroup
)

func main() {
	// On ^C, or SIGTERM handle exit.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Printf("Received %s, exiting.\n", sig.String())
			os.Exit(0)
		}
	}()

	// Find all devices
	devices, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		wg.Add(1)
		go watchDevice(device)
	}

	wg.Wait()
}

func watchDevice(device pcap.Interface) {
	defer wg.Done()

	// Print the device information.
	fmt.Printf("Name: %s\n", device.Name)
	fmt.Printf("Description: %s\n", device.Description)
	fmt.Printf("Devices addresses: %s\n", device.Description)
	for _, address := range device.Addresses {
		fmt.Printf("- IP address: %s\n", address.IP)
		fmt.Printf("- Subnet mask: %s\n", address.Netmask)
	}

	// Open the device.
	fmt.Printf("Opening device %s...\n", device.Name)
	handle, err := pcap.OpenLive(device.Name, 1024, false, 30*time.Second)
	if err != nil {
		log.Printf("opening device %s failed: %v", device.Name, err)
		log.Printf("skipping device %s...", device.Name)
		return
	}
	defer handle.Close()

	// Use the handle as a packet source and process all packets.
	fmt.Printf("Reading packets from %s...\n", device.Name)
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Process the packet.
		if err := printPacketInfo(device.Name, packet); err != nil {
			log.Printf("[%s] printing packet failed: %v", device.Name, err)
		}
		// fmt.Printf("packet: %#v\n", packet)
	}
}

func printPacketInfo(device string, packet gopacket.Packet) error {
	// Check if the packet is an ethernet packet.
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		// Convert to an ethernet packet.
		ethernetPacket, ok := ethernetLayer.(*layers.Ethernet)
		if !ok {
			return errors.New("converting packet to ethernet packet failed")
		}

		fmt.Printf("[%s] source MAC: %s\n", device, ethernetPacket.SrcMAC.String())
		fmt.Printf("[%s] destination MAC: %s\n", device, ethernetPacket.DstMAC.String())
		// Ethernet type is typically IPv4 but could be ARP or other.
		fmt.Printf("[%s] ethernet type: %#v\n\n", device, ethernetPacket.EthernetType)
	}

	// Check if the packet is IP (even though the ether type told us).
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		// Convert to IPv4 packet.
		ip, ok := ipLayer.(*layers.IPv4)
		if !ok {
			return errors.New("converting packet to IPv4 packet failed")
		}

		// IP layer variables:
		// Version (Either 4 or 6)
		// IHL (IP Header Length in 32-bit words)
		// TOS, Length, Id, Flags, FragOffset, TTL, Protocol (TCP?),
		// Checksum, SrcIP, DstIP
		fmt.Printf("[%s] from src IP %s -> dest IP %s\n", device, ip.SrcIP.String(), ip.DstIP.String())
		fmt.Printf("[%s] protocol: %#v\n\n", device, ip.Protocol)
	}

	// Check if the packet is TCP.
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		// Convert to TCP packet.
		tcp, ok := tcpLayer.(*layers.TCP)
		if !ok {
			return errors.New("converting packet to TCP packet failed")
		}

		// TCP layer variables:
		// SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum, Urgent
		// Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
		fmt.Printf("[%s] from src port %s -> dest port %s\n", device, tcp.SrcPort.String(), tcp.DstPort.String())
		fmt.Println("[%s] sequence number: %#v\n", device, tcp.Seq)
		fmt.Printf("[%s] TCP SYN: %t | ACK: %t\n\n", device, tcp.SYN, tcp.ACK)
	}

	// Iterate over all layers, printing out each layer type.
	fmt.Printf("[%s] All packet layers:\n", device)
	for _, layer := range packet.Layers() {
		fmt.Println("- ", layer.LayerType())
	}

	// When iterating through packet.Layers() above,
	// if it lists Payload layer then that is the same as
	// this applicationLayer. applicationLayer contains the payload.
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		fmt.Printf("[%s] payload: %s\n", device, applicationLayer.Payload())

		// Search for a string inside the payload.
		if strings.Contains(string(applicationLayer.Payload()), "HTTP") {
			fmt.Printf("[%s] HTTP found!\n", device)
		}
	}

	// Check for errors.
	if err := packet.ErrorLayer(); err != nil {
		return fmt.Errorf("error decoding part of the packet: %v", err)
	}

	return nil
}
