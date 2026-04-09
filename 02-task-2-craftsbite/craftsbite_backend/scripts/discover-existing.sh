#!/bin/bash
# Discover existing CraftsBite AWS resources
# Run this BEFORE running terraform import to find resource IDs
# Requires: AWS CLI configured, jq

REGION="${1:-ap-southeast-1}"

echo "=== CraftsBite AWS Resource Discovery ==="
echo "Region: $REGION"
echo ""

echo "--- DynamoDB Tables ---"
aws dynamodb list-tables --region "$REGION" --output text
echo ""

echo "--- S3 Buckets (craftsbite-related) ---"
aws s3 ls 2>/dev/null | grep -E "craftsbite|trainee" || echo "(no matching buckets found)"
echo ""

echo "--- SSM Parameters (craftsbite) ---"
aws ssm describe-parameters \
    --region "$REGION" \
    --parameter-type-filters "Type=SecureString" \
    --query "Parameters[?starts_with(Name, '/craftsbite')].Name" \
    --output text 2>/dev/null || echo "(no matching parameters found)"
echo ""

echo "--- Lambda Functions ---"
aws lambda list-functions --region "$REGION" \
    --query "Functions[?contains(FunctionName, 'craftsbite') || contains(FunctionName, 'router') || contains(FunctionName, 'self') || contains(FunctionName, 'management') || contains(FunctionName, 'ops')].[FunctionName, Runtime, Architectures[0], MemorySize, Timeout]" \
    --output table 2>/dev/null || echo "(no matching functions found)"
echo ""

echo "--- API Gateway (HTTP APIs) ---"
aws apigatewayv2 get-apis --region "$REGION" \
    --query "Items.[Name, ApiId, ApiEndpoint, ProtocolType]" \
    --output table 2>/dev/null || echo "(no APIs found)"
echo ""

echo "--- IAM Roles (Lambda-related) ---"
aws iam list-roles --query "Roles[?contains(RoleName, 'craftsbite') || contains(RoleName, 'lambda') || contains(RoleName, 'Lambda')].[RoleName, Arn]" \
    --output table 2>/dev/null || echo "(no matching roles found)"
echo ""

echo "--- CloudWatch Log Groups (Lambda-related) ---"
aws logs describe-log-groups --region "$REGION" \
    --query "logGroups[?contains(logGroupName, '/aws/lambda/')].[logGroupName, retentionInDays]" \
    --output table 2>/dev/null | grep -E "craftsbite|router|self|management|ops" || echo "(no matching log groups found)"
echo ""

echo "=== Discovery Complete ==="
echo ""
echo "Next steps:"
echo "1. Review the output above and note the exact resource names/IDs"
echo "2. Update terraform/environments/prod.tfvars to match existing resource names"
echo "3. Run ./scripts/import-existing.sh to import resources into Terraform state"