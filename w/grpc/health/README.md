




```bash
$ mkdir grpc_health_v1
$ protoc --proto_path=v1 --go_out=grpc_health_v1 --go_opt=paths=source_relative \
  --go-grpc_out=grpc_health_v1 --go-grpc_opt=paths=source_relative health.proto
```