



```bash
$ protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    helloworld.proto
    
$ go run greeter_server/main.go
2023/05/22 02:06:40 server listening at [::]:50051

$ go run greeter_client/main.go --name=Alice
2023/05/22 02:10:50 Greeting: Hello Alice
$ go run greeter_client/main.go --name=Bob
2023/05/22 02:11:02 Greeting: Hello Bob

$ grpcurl -plaintext localhost:50051 list
grpc.reflection.v1alpha.ServerReflection
helloworld.Greeter
$ grpcurl -plaintext localhost:50051 list helloworld.Greeter                                                                      2:18:52 
helloworld.Greeter.SayHello
$ grpcurl -plaintext localhost:50051 list grpc.reflection.v1alpha.ServerReflection
grpc.reflection.v1alpha.ServerReflection.ServerReflectionInfo
$ grpcurl -plaintext -d '{"name": "Alice"}' localhost:50051 helloworld.Greeter/SayHello
{
  "message": "Hello Alice"
}

$ grpcurl -plaintext localhost:50051 list grpc.health.v1.Health
grpc.health.v1.Health.Check
grpc.health.v1.Health.Watch
$ grpcurl -plaintext -d '{"service": "Service001"}' localhost:50051 grpc.health.v1.Health.Check
ERROR:
  Code: NotFound
  Message: unknown service
$ grpcurl -plaintext -d '{"service": "/helloworld.Greeter/SayHello"}' localhost:50051 grpc.health.v1.Health.Check
{
  "status": "SERVING"
}
$ grpcurl -plaintext -d '{"service": "/helloworld.Greeter/SayHello"}' localhost:50051 grpc.health.v1.Health.Watch
$ grpcurl -plaintext -d '{"service": "/helloworld.Greeter/SayHello"}' localhost:50051 grpc.health.v1.Health/Watch
{
  "status": "SERVING"
}
```