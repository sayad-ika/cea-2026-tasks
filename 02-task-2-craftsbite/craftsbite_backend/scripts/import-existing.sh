#!/bin/bash
# Import existing CraftsBite AWS resources into Terraform state
#
# IMPORTANT: Run discover-existing.sh FIRST to find exact resource IDs.
#            Then update the variables below to match YOUR existing resource names.
#
# Usage:
#   1. Edit the variables below to match your AWS resource names
#   2. Run: ./scripts/import-existing.sh
#
# What this does:
#   - Tells Terraform "this resource address already exists in AWS, here's its ID"
#   - Does NOT modify, create, or delete any AWS resources
#   - Only updates the local Terraform state file

set -euo pipefail

# === EDIT THESE TO MATCH YOUR EXISTING RESOURCES ===
# Run discover-existing.sh to find these values

DYNAMODB_TABLE_NAME="trainee-2026-sayad-craftsbite"
S3_BUCKET_NAME="trainee-2026-sayad-craftsbite"

# Lambda function names (from AWS Console or discover script)
# Leave empty to skip importing (Terraform will create new ones)
ROUTER_LAMBDA_NAME=""
GCHAT_ROUTER_LAMBDA_NAME=""
SELF_LAMBDA_NAME=""
MANAGEMENT_LAMBDA_NAME=""
OPS_LAMBDA_NAME=""

# API Gateway ID (from discover script)
API_GATEWAY_ID=""

# IAM Role name (from discover script)
LAMBDA_ROLE_NAME=""

# Terraform working directory
TF_DIR="terraform"

echo "=== CraftsBite Terraform Import ==="
echo "This will import existing AWS resources into Terraform state."
echo "No AWS resources will be modified, created, or deleted."
echo ""

# Check terraform is initialized
if ! terraform -chdir="$TF_DIR" state list > /dev/null 2>&1; then
    echo "Terraform not initialized. Running init..."
    terraform -chdir="$TF_DIR" init -backend=false
fi

IMPORT_COUNT=0
SKIP_COUNT=0

# ---- STATEFUL RESOURCES (Import these to preserve data) ----

echo ""
echo "--- Importing stateful resources ---"

# 1. DynamoDB Table
echo "1/8 Importing DynamoDB table: $DYNAMODB_TABLE_NAME..."
if terraform -chdir="$TF_DIR" import "module.database.aws_dynamodb_table.craftsbite" "$DYNAMODB_TABLE_NAME" 2>/dev/null; then
    ((IMPORT_COUNT++))
else
    echo "   SKIP: Table '$DYNAMODB_TABLE_NAME' not found or already imported"
    ((SKIP_COUNT++))
fi

# 2. S3 Bucket
echo "2/8 Importing S3 bucket: $S3_BUCKET_NAME..."
if terraform -chdir="$TF_DIR" import "module.storage.aws_s3_bucket.deployment" "$S3_BUCKET_NAME" 2>/dev/null; then
    ((IMPORT_COUNT++))
else
    echo "   SKIP: Bucket '$S3_BUCKET_NAME' not found or already imported"
    ((SKIP_COUNT++))
fi

# 3-5. SSM Parameters
for param_name in "DISCORD_BOT_TOKEN" "DISCORD_PUBLIC_KEY" "gchat_service_account_json"; do
    tf_address="module.secrets.aws_ssm_parameter.$(
        case $param_name in
            DISCORD_BOT_TOKEN) echo "discord_bot_token" ;;
            DISCORD_PUBLIC_KEY) echo "discord_public_key" ;;
            gchat_service_account_json) echo "gchat_service_account_json" ;;
        esac
    )"
    ssm_path="/craftsbite/$param_name"
    echo "Importing SSM parameter: $ssm_path..."
    if terraform -chdir="$TF_DIR" import "$tf_address" "$ssm_path" 2>/dev/null; then
        ((IMPORT_COUNT++))
    else
        echo "   SKIP: Parameter '$ssm_path' not found or already imported"
        ((SKIP_COUNT++))
    fi
done

# 6. S3 Bucket Versioning
echo "6/8 Importing S3 bucket versioning..."
if terraform -chdir="$TF_DIR" import "module.storage.aws_s3_bucket_versioning.deployment" "$S3_BUCKET_NAME" 2>/dev/null; then
    ((IMPORT_COUNT++))
else
    echo "   SKIP: Bucket versioning not configured or already imported"
    ((SKIP_COUNT++))
fi

# 7. S3 Public Access Block
echo "7/8 Importing S3 bucket public access block..."
if terraform -chdir="$TF_DIR" import "module.storage.aws_s3_bucket_public_access_block.deployment" "$S3_BUCKET_NAME" 2>/dev/null; then
    ((IMPORT_COUNT++))
else
    echo "   SKIP: Public access block not configured or already imported"
    ((SKIP_COUNT++))
fi

# 8. S3 Encryption
echo "8/8 Importing S3 bucket encryption..."
if terraform -chdir="$TF_DIR" import "module.storage.aws_s3_bucket_server_side_encryption_configuration.deployment" "$S3_BUCKET_NAME" 2>/dev/null; then
    ((IMPORT_COUNT++))
else
    echo "   SKIP: Encryption not configured or already imported"
    ((SKIP_COUNT++))
fi

# ---- STATELESS RESOURCES (Optional) ----

echo ""
echo "--- Stateless resources (optional, can skip to let Terraform create new) ---"

if [ -n "$LAMBDA_ROLE_NAME" ]; then
    echo "Importing IAM role: $LAMBDA_ROLE_NAME..."
    if terraform -chdir="$TF_DIR" import "module.lambda.aws_iam_role.lambda_execution" "$LAMBDA_ROLE_NAME" 2>/dev/null; then
        ((IMPORT_COUNT++))
    else
        echo "   SKIP: IAM role not found or already imported"
        ((SKIP_COUNT++))
    fi
else
    echo "SKIP: No IAM role name provided. Terraform will create a new one."
fi

for func_info in "router:$ROUTER_LAMBDA_NAME" "gchat_router:$GCHAT_ROUTER_LAMBDA_NAME" "self:$SELF_LAMBDA_NAME" "management:$MANAGEMENT_LAMBDA_NAME" "ops:$OPS_LAMBDA_NAME"; do
    tf_name="${func_info%%:*}"
    fn_name="${func_info##*:}"
    if [ -n "$fn_name" ]; then
        echo "Importing Lambda function: $fn_name..."
        if terraform -chdir="$TF_DIR" import "module.lambda.aws_lambda_function.$tf_name" "$fn_name" 2>/dev/null; then
            ((IMPORT_COUNT++))
        else
            echo "   SKIP: Function '$fn_name' not found or already imported"
            ((SKIP_COUNT++))
        fi
    fi
done

if [ -n "$API_GATEWAY_ID" ]; then
    echo "Importing API Gateway: $API_GATEWAY_ID..."
    if terraform -chdir="$TF_DIR" import "module.api_gateway.aws_apigatewayv2_api.craftsbite" "$API_GATEWAY_ID" 2>/dev/null; then
        ((IMPORT_COUNT++))
    else
        echo "   SKIP: API Gateway not found or already imported"
        ((SKIP_COUNT++))
    fi
else
    echo "SKIP: No API Gateway ID provided. Terraform will create a new one."
fi

# ---- SUMMARY ----

echo ""
echo "=== Import Summary ==="
echo "Imported: $IMPORT_COUNT resources"
echo "Skipped:  $SKIP_COUNT resources"
echo ""
echo "Next steps:"
echo "1. Run: terraform -chdir=terraform plan -var-file=environments/prod.tfvars"
echo "   Review the plan carefully. Look for:"
echo "   - '+ create' for new resources (Lambdas, IAM, API Gateway) = GOOD"
echo "   - '~ update in-place' for config drift = usually OK"
echo "   - '- destroy' on DynamoDB table = BAD, stop immediately"
echo "2. If plan looks safe, run: terraform -chdir=terraform apply -var-file=environments/prod.tfvars"