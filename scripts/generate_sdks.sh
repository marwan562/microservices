#!/bin/bash
set -e

# This script generates SDKs for Node, Go, and Python from the OpenAPI spec.

# Ensure we are in the root directory
cd "$(dirname "$0")/.."

SWAGGER_FILE="generated/openapi/fintech.swagger.json"

if [ ! -f "$SWAGGER_FILE" ]; then
    echo "Swagger file not found. Run scripts/generate_openapi.sh first."
    exit 1
fi

echo "Generating SDKs..."

# --- Node.js ---
echo "Generating Node.js SDK..."
mkdir -p ../fintech-sdk-node/src/generated
openapi-generator-cli generate -i "$SWAGGER_FILE" -g typescript-axios -o ../fintech-sdk-node/src/generated --additional-properties=npmName=@sapliyio/fintech-node-generated,supportsES6=true

# --- Go ---
echo "Generating Go SDK..."
mkdir -p ../fintech-sdk-go/generated
openapi-generator-cli generate -i "$SWAGGER_FILE" -g go -o ../fintech-sdk-go/generated \
  --additional-properties=packageName=generated,enumClassPrefix=true \
  --git-host github.com --git-user-id sapliy --git-repo-id fintech-sdk-go

# Fix GIT_USER_ID/GIT_REPO_ID placeholders in generated tests
sed -i '' 's/github.com\/GIT_USER_ID\/GIT_REPO_ID/github.com\/sapliy\/fintech-sdk-go\/generated/g' ../fintech-sdk-go/generated/test/*.go

# --- Python ---
echo "Generating Python SDK..."
mkdir -p ../fintech-sdk-python/sapliy_fintech/generated
openapi-generator-cli generate -i "$SWAGGER_FILE" -g python -o ../fintech-sdk-python/sapliy_fintech/generated --additional-properties=packageName=sapliy_fintech.generated

echo "SDK generation complete."
