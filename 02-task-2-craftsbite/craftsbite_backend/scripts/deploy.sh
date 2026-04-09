#!/bin/bash
# Deploy CraftsBite infrastructure
# Builds Lambda zips, uploads to S3, and applies Terraform
# Usage: ./scripts/deploy.sh [environment]
#   environment: dev (default), staging, or prod
set -euo pipefail

ENVIRONMENT="${1:-dev}"

echo "=== Deploying CraftsBite (${ENVIRONMENT}) ==="

# Step 1: Build Lambda zips
echo ""
echo "--- Step 1: Building Lambda functions ---"
./scripts/build.sh

# Step 2: Initialize Terraform (if not already initialized)
echo ""
echo "--- Step 2: Initializing Terraform ---"
terraform -chdir=terraform init

# Step 3: Plan
echo ""
echo "--- Step 3: Planning Terraform changes ---"
terraform -chdir=terraform plan \
  -var-file="environments/${ENVIRONMENT}.tfvars" \
  -out="tfplan"

# Step 4: Apply
echo ""
echo "--- Step 4: Applying Terraform changes ---"
terraform -chdir=terraform apply "tfplan"

echo ""
echo "=== Deployment complete ==="