# Import existing CraftsBite AWS resources into Terraform state
#
# IMPORTANT: Run discover-existing.ps1 FIRST to find exact resource IDs.
#            Then update the variables below to match YOUR existing resource names.
#
# Usage:
#   1. Edit the variables below to match your AWS resource names
#   2. Run: .\scripts\import-existing.ps1
#
# What this does:
#   - Tells Terraform "this resource address already exists in AWS, here's its ID"
#   - Does NOT modify, create, or delete any AWS resources
#   - Only updates the local Terraform state file
#
# After importing:
#   - Run `terraform plan` to see the diff between current state and desired config
#   - Resources you imported will show "no changes" or minor config drift
#   - New resources (Lambdas, IAM, API Gateway) will show as "create"

param(
    # === EDIT THESE TO MATCH YOUR EXISTING RESOURCES ===
    # Run discover-existing.ps1 to find these values

    [string]$DynamoDBTableName = "trainee-2026-sayad-craftsbite",
    [string]$S3BucketName      = "trainee-2026-sayad-craftsbite",

    # Lambda function names (from AWS Console or discover script)
    [string]$RouterLambdaName       = "",
    [string]$GChatRouterLambdaName   = "",
    [string]$SelfLambdaName         = "",
    [string]$ManagementLambdaName   = "",
    [string]$OpsLambdaName          = "",

    # API Gateway ID (from discover script)
    [string]$ApiGatewayId           = "",

    # IAM Role name (from discover script)
    [string]$LambdaRoleName         = "",

    # Terraform working directory (relative to project root)
    [string]$TerraformDir          = "terraform"
)

$ErrorActionPreference = "Stop"
$tfChdir = "-chdir=$TerraformDir"
$importCount = 0
$skipCount = 0

Write-Host "=== CraftsBite Terraform Import ===" -ForegroundColor Cyan
Write-Host "This will import existing AWS resources into Terraform state."
Write-Host "No AWS resources will be modified, created, or deleted."
Write-Host ""

# Check that terraform is initialized
Write-Host "Checking Terraform initialization..." -ForegroundColor Yellow
$initCheck = terraform $tfChdir state list 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "Terraform not initialized. Running init..." -ForegroundColor Yellow
    terraform $tfChdir init -backend=false
}

# ---- STATEFUL RESOURCES (Import these to preserve data) ----

Write-Host ""
Write-Host "--- Importing stateful resources ---" -ForegroundColor Green

# 1. DynamoDB Table
Write-Host "1/8 Importing DynamoDB table: $DynamoDBTableName..." -ForegroundColor Cyan
try {
    terraform $tfChdir import "module.database.aws_dynamodb_table.craftsbite" $DynamoDBTableName 2>&1 | Write-Host
    $importCount++
} catch {
    Write-Host "   SKIP: Table '$DynamoDBTableName' not found or already imported" -ForegroundColor Yellow
    $skipCount++
}

# 2. S3 Bucket
Write-Host "2/8 Importing S3 bucket: $S3BucketName..." -ForegroundColor Cyan
try {
    terraform $tfChdir import "module.storage.aws_s3_bucket.deployment" $S3BucketName 2>&1 | Write-Host
    $importCount++
} catch {
    Write-Host "   SKIP: Bucket '$S3BucketName' not found or already imported" -ForegroundColor Yellow
    $skipCount++
}

# 3. SSM Parameters
Write-Host "3/8 Importing SSM parameter: /craftsbite/DISCORD_BOT_TOKEN..." -ForegroundColor Cyan
try {
    terraform $tfChdir import "module.secrets.aws_ssm_parameter.discord_bot_token" "/craftsbite/DISCORD_BOT_TOKEN" 2>&1 | Write-Host
    $importCount++
} catch {
    Write-Host "   SKIP: Parameter not found or already imported" -ForegroundColor Yellow
    $skipCount++
}

Write-Host "4/8 Importing SSM parameter: /craftsbite/DISCORD_PUBLIC_KEY..." -ForegroundColor Cyan
try {
    terraform $tfChdir import "module.secrets.aws_ssm_parameter.discord_public_key" "/craftsbite/DISCORD_PUBLIC_KEY" 2>&1 | Write-Host
    $importCount++
} catch {
    Write-Host "   SKIP: Parameter not found or already imported" -ForegroundColor Yellow
    $skipCount++
}

Write-Host "5/8 Importing SSM parameter: /craftsbite/gchat_service_account_json..." -ForegroundColor Cyan
try {
    terraform $tfChdir import "module.secrets.aws_ssm_parameter.gchat_service_account_json" "/craftsbite/gchat_service_account_json" 2>&1 | Write-Host
    $importCount++
} catch {
    Write-Host "   SKIP: Parameter not found or already imported" -ForegroundColor Yellow
    $skipCount++
}

# ---- S3 BUCKET SUB-RESOURCES (Optional, import if they exist) ----

Write-Host ""
Write-Host "--- S3 bucket sub-resources (import if configured) ---" -ForegroundColor Green
Write-Host "Checking S3 bucket versioning..." -ForegroundColor Cyan
$versioning = aws s3api get-bucket-versioning --bucket $S3BucketName 2>$null
if ($versioning -match "Enabled") {
    Write-Host "6/8 Importing S3 bucket versioning..." -ForegroundColor Cyan
    terraform $tfChdir import "module.storage.aws_s3_bucket_versioning.deployment" $S3BucketName 2>&1 | Write-Host
    $importCount++
} else {
    Write-Host "6/8 SKIP: Bucket versioning not enabled, Terraform will create it" -ForegroundColor Yellow
    $skipCount++
}

Write-Host "7/8 Importing S3 bucket public access block..." -ForegroundColor Cyan
try {
    terraform $tfChdir import "module.storage.aws_s3_bucket_public_access_block.deployment" $S3BucketName 2>&1 | Write-Host
    $importCount++
} catch {
    Write-Host "   SKIP: Public access block not configured, Terraform will create it" -ForegroundColor Yellow
    $skipCount++
}

Write-Host "8/8 Importing S3 bucket encryption..." -ForegroundColor Cyan
try {
    terraform $tfChdir import "module.storage.aws_s3_bucket_server_side_encryption_configuration.deployment" $S3BucketName 2>&1 | Write-Host
    $importCount++
} catch {
    Write-Host "   SKIP: Encryption not configured, Terraform will create it" -ForegroundColor Yellow
    $skipCount++
}

# ---- STATELESS RESOURCES (Optional, can also let Terraform create new) ----

Write-Host ""
Write-Host "--- Stateless resources (Lambdas, IAM, API Gateway) ---" -ForegroundColor Green

if ($LambdaRoleName -ne "") {
    Write-Host "Importing IAM role: $LambdaRoleName..." -ForegroundColor Cyan
    try {
        terraform $tfChdir import "module.lambda.aws_iam_role.lambda_execution" $LambdaRoleName 2>&1 | Write-Host
        $importCount++
    } catch {
        Write-Host "   SKIP: IAM role '$LambdaRoleName' not found or already imported" -ForegroundColor Yellow
        $skipCount++
    }
} else {
    Write-Host "SKIP: No IAM role name provided. Terraform will create a new one." -ForegroundColor Yellow
    Write-Host "      Set -LambdaRoleName if you want to import your existing role." -ForegroundColor Yellow
}

if ($RouterLambdaName -ne "") {
    Write-Host "Importing Lambda function: $RouterLambdaName..." -ForegroundColor Cyan
    try {
        terraform $tfChdir import "module.lambda.aws_lambda_function.router" $RouterLambdaName 2>&1 | Write-Host
        $importCount++
    } catch {
        Write-Host "   SKIP: Function '$RouterLambdaName' not found or already imported" -ForegroundColor Yellow
        $skipCount++
    }
} else {
    Write-Host "SKIP: No router Lambda name provided. Terraform will create a new one." -ForegroundColor Yellow
}

if ($ApiGatewayId -ne "") {
    Write-Host "Importing API Gateway: $ApiGatewayId..." -ForegroundColor Cyan
    try {
        terraform $tfChdir import "module.api_gateway.aws_apigatewayv2_api.craftsbite" $ApiGatewayId 2>&1 | Write-Host
        $importCount++
    } catch {
        Write-Host "   SKIP: API Gateway '$ApiGatewayId' not found or already imported" -ForegroundColor Yellow
        $skipCount++
    }
} else {
    Write-Host "SKIP: No API Gateway ID provided. Terraform will create a new one." -ForegroundColor Yellow
}

# ---- SUMMARY ----

Write-Host ""
Write-Host "=== Import Summary ===" -ForegroundColor Cyan
Write-Host "Imported: $importCount resources" -ForegroundColor Green
Write-Host "Skipped:  $skipCount resources" -ForegroundColor Yellow
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Green
Write-Host "1. Run: terraform -chdir=terraform plan -var-file=environments/prod.tfvars" -ForegroundColor White
Write-Host "   Review the plan carefully. Look for:" -ForegroundColor White
Write-Host "   - '+ create' for new resources (Lambdas, IAM, API Gateway) = GOOD" -ForegroundColor White
Write-Host "   - '~ update in-place' for config drift = usually OK" -ForegroundColor White
Write-Host "   - '- destroy' on DynamoDB table = BAD, stop immediately" -ForegroundColor Red
Write-Host "2. If plan looks safe, run: terraform -chdir=terraform apply -var-file=environments/prod.tfvars" -ForegroundColor White