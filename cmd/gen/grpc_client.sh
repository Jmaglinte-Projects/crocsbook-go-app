#! /bin/bash -e 

SRC=~/Documents/Github/Jmaglinte-Projects/crocsbook-go-app/infra/grpc/protos
OUTPUT=~/Documents/Github/Jmaglinte-Projects/crocsbook-api-client/api

# PROTOC_GEN_TS_PATH=/Users/jaffy/.nvm/versions/node/v24.2.0/bin/protoc-gen-ts_proto
 
rm -rf $OUTPUT
mkdir $OUTPUT


# https://github.com/stephenh/ts-proto
protoc --plugin=./node_modules/.bin/protoc-gen-ts_proto \
  --ts_proto_out=$OUTPUT \
  --ts_proto_opt=outputServices=grpc-js \
  -I $SRC \
   "$SRC"/*.proto
 