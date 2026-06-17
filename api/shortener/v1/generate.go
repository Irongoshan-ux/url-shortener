package v1

//go:generate protoc -I . --go_out=../../../pkg/shortener/v1 --go_opt=paths=source_relative --go-grpc_out=../../../pkg/shortener/v1 --go-grpc_opt=paths=source_relative shortener.proto
