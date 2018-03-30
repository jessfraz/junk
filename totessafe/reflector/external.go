package reflector

import (
	"fmt"
	"net"
	"strconv"

	"github.com/kubicorn/kubicorn/pkg/logger"
	"google.golang.org/grpc"
)

func ConcurrentExternalListenAndServe(externalPort int) chan error {
	ch := make(chan error)
	go func() {
		port := fmt.Sprintf(":%s", strconv.Itoa(externalPort))
		listener, err := net.Listen("tcp", port)
		if err != nil {
			ch <- fmt.Errorf("Unable to open TCP socket to listen on: %v", err)
			return
		}
		server := grpc.NewServer()
		RegisterReflectorExternalServer(server, NewReflectorExternalServer())
		logger.Always("Starting [EXTERNAL] gRPC server...")
		if err := server.Serve(listener); err != nil {
			ch <- fmt.Errorf("Failed to start serving gRPC service: %v", err)
			return
		}
	}()
	return ch
}
