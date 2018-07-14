package types

import (
	"github.com/google/gopacket/layers"
)

// PacketBlob holds the object structure for a packet blob.
type PacketBlob struct {
	LayerType string `json:"type,omitempty"`
	Device    string `json:"device,omitempty"`

	Payload   string `json:"payload,omitempty"`
	FoundHTTP bool   `json:"foundHTTP,omitempty"`

	// Ethernet options.
	SrcMAC       string              `json:"srcMAC,omitempty"`
	DstMAC       string              `json:"dstMAC,omitempty"`
	EthernetType layers.EthernetType `json:"ethernetType,omitempty"`

	// IPv4 options.
	SrcIP      string            `json:"srcIP,omitempty"`
	DstIP      string            `json:"dstIP,omitempty"`
	IPProtocol layers.IPProtocol `json:"ipProtocol,omitempty"`

	// TCP options.
	SrcPort        string `json:"srcPort,omitempty"`
	DstPort        string `json:"dstPort,omitempty"`
	SequenceNumber uint32 `json:"sequenceNumber,omitempty"`
	SYN            bool   `json:"syn,omitempty"`
	ACK            bool   `json:"ack,omitempty"`

	// Raw layer info.
	Layers interface{} `json:"layers,omitempty"`
}

// ProcBlob holds the object structure for a /proc blob.
type ProcBlob struct {
	PID     int      `json:"pid,omitempty"`
	Env     []string `json:"env,omitempty"`
	Cmdline []string `json:"cmdline,omitempty"`
	Cwd     string   `json:"cwd,omitempty"`
	Exe     string   `json:"exe,omitempty"`
}
