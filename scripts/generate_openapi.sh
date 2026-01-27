#!/bin/bash
set -e

# This script generates OpenAPI v2 specs from Protocol Buffers using a Docker container.
# It mimics what happens in the GitHub Action.

# Ensure we are in the root directory
cd "$(dirname "$0")/.."

echo "Generating OpenAPI specs..."

# Create output directory
mkdir -p generated/openapi

# Run buf or protoc via Docker to generate openapi.json
# Using a widely used image for convenience (e.g., bufbuild/buf or similar via protoc)
# For simplicity in this environment without buf installed, we'll try to use a standard protoc image 
# that has grpc-gateway logic or similar. 
# Actually, the plan specified using 'openapi-generator-cli' later, but first we need the Spec.
# We will use a dockerized protoc command.

docker run --rm -v $(pwd):/defs -w /defs namely/protoc-all:1.51_2 \
    -d proto \
    -l openapi \
    --with-gateway \
    -o generated/openapi

echo "OpenAPI specs generated in generated/openapi/"
