package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/jessfraz/paws/moarpackets/types"
	"github.com/jessfraz/paws/totessafe/reflector"
)

func watchDevice(client *reflector.InternalReflectorClient, device pcap.Interface) {
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
		data, err := getPacketInfo(device.Name, packet)
		if err != nil {
			log.Printf("[%s] printing packet failed: %v", device.Name, err)
			continue
		}

		b, err := json.Marshal(data)
		if err != nil {
			log.Printf("[%s] marshal packet data failed: %v", device.Name, err)
			continue
		}

		blob := &reflector.PawsBlob{
			Data: string(b),
		}
		if _, err = client.Client.Set(context.TODO(), blob); err != nil {
			log.Printf("[%s] sending packet data to totessafe client failed: %v", device.Name, err)
			continue
		}
	}
}

func getPacketInfo(device string, packet gopacket.Packet) (*types.PacketBlob, error) {
	d := &types.PacketBlob{
		Device: device,
	}

	// Check if the packet is an ethernet packet.
	ethernetLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethernetLayer != nil {
		d.LayerType = "ethernet"

		// Convert to an ethernet packet.
		ethernetPacket, ok := ethernetLayer.(*layers.Ethernet)
		if !ok {
			return nil, errors.New("converting packet to ethernet packet failed")
		}

		d.SrcMAC = ethernetPacket.SrcMAC.String()
		d.DstMAC = ethernetPacket.DstMAC.String()
		d.EthernetType = ethernetPacket.EthernetType
	}

	// Check if the packet is IP (even though the ether type told us).
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		d.LayerType = "IPv4"

		// Convert to IPv4 packet.
		ip, ok := ipLayer.(*layers.IPv4)
		if !ok {
			return nil, errors.New("converting packet to IPv4 packet failed")
		}

		// IP layer variables:
		// Version (Either 4 or 6)
		// IHL (IP Header Length in 32-bit words)
		// TOS, Length, Id, Flags, FragOffset, TTL, Protocol (TCP?),
		// Checksum, SrcIP, DstIP
		d.SrcIP = ip.SrcIP.String()
		d.DstIP = ip.DstIP.String()
		d.IPProtocol = ip.Protocol
	}

	// Check if the packet is TCP.
	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	if tcpLayer != nil {
		d.LayerType = "TCP"

		// Convert to TCP packet.
		tcp, ok := tcpLayer.(*layers.TCP)
		if !ok {
			return nil, errors.New("converting packet to TCP packet failed")
		}

		// TCP layer variables:
		// SrcPort, DstPort, Seq, Ack, DataOffset, Window, Checksum, Urgent
		// Bool flags: FIN, SYN, RST, PSH, ACK, URG, ECE, CWR, NS
		d.SrcPort = tcp.SrcPort.String()
		d.DstPort = tcp.DstPort.String()
		d.SequenceNumber = tcp.Seq
		d.SYN = tcp.SYN
		d.ACK = tcp.ACK
	}

	// Iterate over all layers, printing out each layer type.
	d.Layers = packet.Layers()

	// When iterating through packet.Layers() above,
	// if it lists Payload layer then that is the same as
	// this applicationLayer. applicationLayer contains the payload.
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		d.Payload = string(applicationLayer.Payload())

		// Search for a string inside the payload.
		if strings.Contains(string(applicationLayer.Payload()), "HTTP") {
			d.FoundHTTP = true
		}
	}

	// Check for errors.
	if err := packet.ErrorLayer(); err != nil {
		return nil, fmt.Errorf("error decoding part of the packet: %v", err)
	}

	return d, nil
}
