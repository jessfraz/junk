package client

import (
	"fmt"

	"github.com/kubicorn/kubicorn/pkg/logger"
	"google.golang.org/grpc"
)

type InternalReflectorClient struct {
	hostname string
	port     int
	Client   ReflectorInternalClient
}

func NewInternalReflectorClient(hostname string, port int) *InternalReflectorClient {
	return &InternalReflectorClient{
		hostname: hostname,
		port:     port,
	}
}

func (c *InternalReflectorClient) dialable() string {
	return fmt.Sprintf("%s:%d", c.hostname, c.port)
}

func (c *InternalReflectorClient) Connect() error {
	logger.Info("Connecting to gRPC server [%s:%d]", c.hostname, c.port)
	connection, err := grpc.Dial(c.dialable(), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("Unable to connect to host [%s] with error message: %v", c.dialable(), err)
	}
	//defer connection.Close()
	client := NewReflectorInternalClient(connection)
	c.Client = client
	return nil
}
