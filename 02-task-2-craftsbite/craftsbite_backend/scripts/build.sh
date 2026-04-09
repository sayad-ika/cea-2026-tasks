#!/bin/bash
# Build all CraftsBite Lambda functions for ARM64 (Graviton)
# Output: dist/{lambda_name}/{lambda_name}.zip
set -euo pipefail

LAMBDAS=("router" "gchat-router" "self" "management" "ops")
GOOS=linux
GOARCH=arm64
CGO_ENABLED=0

echo "=== Building CraftsBite Lambda functions (GOOS=$GOOS, GOARCH=$GOARCH) ==="

for lambda in "${LAMBDAS[@]}"; do
  echo ""
  echo "--- Building $lambda ---"

  # Determine the source directory
  # gchat-router has multiple .go files in its directory
  src_dir="./cmd/$lambda"

  if [ ! -d "$src_dir" ]; then
    echo "ERROR: Source directory $src_dir not found"
    exit 1
  fi

  # Create output directory
  mkdir -p "dist/$lambda"

  # Build the binary
  echo "Compiling..."
  GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=$CGO_ENABLED \
    go build -o "dist/$lambda/bootstrap" "./$src_dir"

  # Create zip (remove old zip first)
  rm -f "dist/$lambda/$lambda.zip"
  cd "dist/$lambda"
  zip "$lambda.zip" bootstrap
  cd ../..

  # Clean up binary
  rm -f "dist/$lambda/bootstrap"

  echo "OK: dist/$lambda/$lambda.zip"
done

echo ""
echo "=== All builds complete ==="
echo "Run 'terraform -chdir=terraform apply' to deploy"