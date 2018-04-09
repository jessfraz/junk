package devops

import (
	"github.com/jessfraz/k8s-aks-dns-ingress/azure"
	context "golang.org/x/net/context"
)

var _ DNSRecordServer = DNSRecordService{}

// DNSRecordService implements a DNSRecordServer.
type DNSRecordService struct {
	azAuth *azure.Authentication
}

// NewDNSRecordServer returns a new DNSRecordServer struct.
func NewDNSRecordServer(azAuth *azure.Authentication) *DNSRecordService {
	return &DNSRecordService{azAuth: azAuth}
}

// Add creates a new dns record for an ingress loadbalancer.
func (d DNSRecordService) Add(ctx context.Context, req *DNSRecordRequest) (*Status, error) {
	return nil, nil
}

// Delete deletes a dns record for an ingress loadbalancer.
func (d DNSRecordService) Delete(ctx context.Context, req *DNSRecordRequest) (*Status, error) {
	return nil, nil
}

// Update updates a dns record for an ingress loadbalancer.
func (d DNSRecordService) Update(ctx context.Context, req *DNSRecordRequest) (*Status, error) {
	return nil, nil
}
