# External

The transaction from a remote client (your laptop) to the reflector in Kubernetes.

### Client

Your laptop, or the client connecting to the reflector.

Example client code:


``` go
client := reflector.NewInternalReflectorClient("totessafe-deployment", 14410)
err := client.Connect()
if err != nil {
    panic(err.Error())
}
blob := &PawsBlob{
    Data: "some string",
}
_, err = client.Client.Set(context.TODO(), blob)
if err != nil {
    panic(err.Error())
}
```

### Server

One of the two processes that will serve as a gRPC server in the reflector.

# Internal

The transaction from a spoofed `/pause` container to the reflector in Kubernetes.

### Client

The spoofed `/pause` container connecting to the reflector.

Example client code:

```go
// This would be the public IP of the Kubernetes Ingress
client := reflector.NewExternalReflectorClient("12.34.56.78", 14411)
err := client.Connect()
if err != nil {
    panic(err.Error())
}
blob, err := client.Client.Get(context.TODO(), &RequestType{})
if err != nil {
    panic(err.Error())
}
fmt.Println(blob.Data)
```

### Server

One of the two processes that will serve as a gRPC server in the reflector.
