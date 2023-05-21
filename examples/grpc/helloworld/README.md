



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

$ grpcurl -plaintext -d '{"name": "Alice"}' localhost:50051 helloworld.Greeter/SayHello
{
  "message": "Hello Alice"
}
```