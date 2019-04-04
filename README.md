# Base
## A sensible, default gRPC server.

Configuring interceptors and options on a gRPC server is relatively simple, but requires a fair bit of boilerplate. Base prevents configuration drift as the number of developers and services grow. 

A marginal improvment...

```go
import (
    "github.com/alee792/base"
    "go.uber.org/zap"
    "google.golang.org/grpc"
    pb "fake/hello.pb.go"
    hello "fake/hello"
)
...
l := zap.NewExample()
s, err := base.NewServer(
    base.Log(l),
    base.TLS("/path/to/cert", "/path/to/key"),
    base.Bundle(
        grpc.MaxConcurrentStreams(25),
        grpc.MaxMsgSize(64000),
    ),
)
if err != nil {
    l.Fatal(err)
}

pb.RegisterHelloServer(s, hello.New())
if err := s.ListenAndServe(":8443"); err != nil {
    l.Fatal(err)
}

...
```