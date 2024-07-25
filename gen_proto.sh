#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Define variables
PROTO_DIR="./proto"

# Check if protoc is installed
if ! command -v protoc &> /dev/null
then
    echo "protoc could not be found. Please install Protocol Buffers Compiler."
    exit 1
fi

# Check if required Go plugins are installed
if ! command -v protoc-gen-go &> /dev/null || ! command -v protoc-gen-go-grpc &> /dev/null
then
    echo "Required Go plugins not found. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create proto directory if it doesn't exist
mkdir -p $PROTO_DIR

# Find all .proto files recursively
proto_files=$(find $PROTO_DIR -name "*.proto")
if [ -z "$proto_files" ]; then
    echo "Error: No .proto files found in $PROTO_DIR"
    exit 1
fi

# Generate Go code for each .proto file
for proto_file in $proto_files
do
    echo "Generating Go code from $(basename $proto_file)..."
    
    # Get the directory of the current .proto file
    proto_dir=$(dirname "$proto_file")
    
    protoc --proto_path=$PROTO_DIR \
           --go_out=$proto_dir --go_opt=paths=source_relative \
           --go-grpc_out=$proto_dir --go-grpc_opt=paths=source_relative \
           $proto_file
done

echo "Go code generation complete."