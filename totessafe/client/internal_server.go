package client

import (
	"golang.org/x/net/context"
)

var _ ReflectorInternalServer = InternalReflectorServer{}

type InternalReflectorServer struct {
	// TODO define parts here
}

func NewInternalReflectorServer() *InternalReflectorServer {
	return &InternalReflectorServer{}
}

// Set is used to set new data to a running gRPC service. Set works like an upsert and will update or create
// as needed.
func (t InternalReflectorServer) Set(ctx context.Context, request *PawsBlob) (*ReturnType, error) {

	//
	//
	// TODO Logic with *PawsBlob
	//
	//

	response := &ReturnType{}
	return response, nil
}
