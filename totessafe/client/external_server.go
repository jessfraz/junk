package client

import (
	//"github.com/kris-nova/terraformctl/parser"
	//"github.com/kris-nova/terraformctl/storage"
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

	//
	//
	// TODO Logic with *PawsBlob
	//
	//

	response := &PawsBlob{}
	return response, nil
}
