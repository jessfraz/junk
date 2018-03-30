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

	// Return it
	return blob, nil
}
