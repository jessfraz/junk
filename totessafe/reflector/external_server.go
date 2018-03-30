package reflector

import (
	"golang.org/x/net/context"
)

var _ ReflectorExternalServer = ExternalReflectorServer{}

type ExternalReflectorServer struct {
	// TODO define parts here
}

func NewReflectorExternalServer() *ExternalReflectorServer {
	return &ExternalReflectorServer{}
}

// Get is used to retrieve data from a running gRPC service.
func (t ExternalReflectorServer) Get(ctx context.Context, requestType *RequestType) (*PawsBlob, error) {

	// Pop a blob off the queue (or nil)
	blob := popBlob()

	// Handle nil case so we don't get error: "proto: Marshal called with nil"
	if blob == nil {
		blob = &PawsBlob{}
	}

	// Return it
	return blob, nil
}
