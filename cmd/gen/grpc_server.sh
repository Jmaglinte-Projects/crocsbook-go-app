#! /bin/bash -e 

cd infra/grpc/protos
rm -rf ../lib
mkdir ../lib

# Generate Go code from protobuf definitions
protoc --go_out=../lib \
       --go_opt=paths=source_relative \
       --go-grpc_out=../lib \
       --go-grpc_opt=paths=source_relative \
       *.proto

protoc -I . --grpc-gateway_out ../lib \
    --grpc-gateway_opt paths=source_relative \
    *.proto