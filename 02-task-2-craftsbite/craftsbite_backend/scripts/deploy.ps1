# Deploy CraftsBite infrastructure
# Builds Lambda zips and applies Terraform
# Usage: .\scripts\deploy.ps1 [environment]
#   environment: dev (default), staging, or prod

param(
    [string]$Environment = "dev"
)

$ErrorActionPreference = "Stop"

Write-Host "=== Deploying CraftsBite ($Environment) ===" -ForegroundColor Cyan

# Step 1: Build Lambda zips
Write-Host ""
Write-Host "--- Step 1: Building Lambda functions ---" -ForegroundColor Yellow
& ".\scripts\build.ps1"

if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed, aborting deployment." -ForegroundColor Red
    exit 1
}

# Step 2: Initialize Terraform (if not already initialized)
Write-Host ""
Write-Host "--- Step 2: Initializing Terraform ---" -ForegroundColor Yellow
terraform -chdir=terraform init

# Step 3: Plan
Write-Host ""
Write-Host "--- Step 3: Planning Terraform changes ---" -ForegroundColor Yellow
terraform -chdir=terraform plan `
    -var-file="environments/$Environment.tfvars" `
    -out="tfplan"

# Step 4: Apply
Write-Host ""
Write-Host "--- Step 4: Applying Terraform changes ---" -ForegroundColor Yellow
terraform -chdir=terraform apply "tfplan"

Write-Host ""
Write-Host "=== Deployment complete ===" -ForegroundColor Green