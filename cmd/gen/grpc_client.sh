#! /bin/bash -e 

SERVER_REPO=~/Documents/Github/Jmaglinte-Projects/room-scheduler-go-app
# CLIENT_REPO=~/Documents/Github/Jmaglinte-Projects/room-scheduler-rrfm
CLIENT_REPO=~/Documents/Github/Jmaglinte-Projects/crocsbook-rrfm
# OUTPUT=~/Documents/Github/Jmaglinte-Projects/room-scheduler-rrfm/app/lib/api
OUTPUT=~/Documents/Github/Jmaglinte-Projects/crocsbook-rrfm/app/lib/api

PROTOC_GEN_TS_PATH=/Users/jaffy/.nvm/versions/node/v24.2.0/bin/protoc-gen-ts_proto

cd internal/infra/grpc/proto
rm -rf $CLIENT_REPO/app/lib/api
mkdir $CLIENT_REPO/app/lib/api
 
protoc --plugin=protoc-gen-ts_proto=/Users/jaffy/.nvm/versions/node/v24.2.0/bin/protoc-gen-ts_proto \
  --ts_proto_out=$OUTPUT \
  --ts_proto_opt=outputServices=grpc-js \
  -I $SERVER_REPO/internal/infra/grpc/proto \
  *.proto
 