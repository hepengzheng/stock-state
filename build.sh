#!/bin/bash


PROTO_DIR="./proto"
API_DIR="./api"

function generate() {
    for proto_file in "$PROTO_DIR"/*.proto; do
        proto_base=$(basename "$proto_file" .proto)
        target_dir="$API_DIR/$proto_base"
        mkdir -p "$target_dir"
        protoc --proto_path="$PROTO_DIR" \
          --go_out="$target_dir" --go_opt=paths=source_relative \
          --go-grpc_out="$target_dir" --go-grpc_opt=paths=source_relative \
          "$proto_file"
        echo "Generated $target_dir/${proto_base}.pb.go"
    done
    echo "gRPC code generation completed successfully!"
}

generate
