# External

The transaction from a remote client (your laptop) to the reflector in Kubernetes.

### Client

Your laptop, or the client connecting to the reflector.

### Server

One of the two processes that will serve as a gRPC server in the reflector.

# Internal

The transaction from a spoofed `/pause` container to the reflector in Kubernetes.

### Client

The spoofed `/pause` container connecting to the reflector.

### Server

One of the two processes that will serve as a gRPC server in the reflector.