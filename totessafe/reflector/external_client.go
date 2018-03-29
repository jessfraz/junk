package reflector

import (
"fmt"
"github.com/kubicorn/kubicorn/pkg/logger"
"google.golang.org/grpc"
)

type ExternalReflectorClient struct {
	hostname string
	port     int
	Client   ReflectorExternalClient
}


func NewExternalReflectorClient(hostname string, port int) *ExternalReflectorClient {
	return &ExternalReflectorClient{
		hostname: hostname,
		port:     port,
	}
}

func (c *ExternalReflectorClient) dialable() string {
	return fmt.Sprintf("%s:%d", c.hostname, c.port)
}

func (c *ExternalReflectorClient) Connect() error {
	logger.Info("Connecting to gRPC server [%s:%d]", c.hostname, c.port)
	connection, err := grpc.Dial(c.dialable(), grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("Unable to connect to host [%s] with error message: %v", c.dialable(), err)
	}
	//defer connection.Close()
	client := NewReflectorExternalClient(connection)
	c.Client = client
	return nil
}

