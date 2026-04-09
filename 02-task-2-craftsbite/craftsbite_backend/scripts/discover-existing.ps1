# Discover existing CraftsBite AWS resources
# Run this BEFORE running terraform import to find resource IDs
# Requires: AWS CLI configured, jq (optional, for JSON parsing)

param(
    [string]$Region = "ap-southeast-1"
)

$ErrorActionPreference = "Continue"

Write-Host "=== CraftsBite AWS Resource Discovery ===" -ForegroundColor Cyan
Write-Host "Region: $Region`n" -ForegroundColor Yellow

# 1. DynamoDB Tables
Write-Host "--- DynamoDB Tables ---" -ForegroundColor Yellow
$dynamoTables = aws dynamodb list-tables --region $Region --output text 2>$null
Write-Host $dynamoTables
Write-Host ""

# 2. S3 Buckets (filter for craftsbite-related)
Write-Host "--- S3 Buckets (craftsbite-related) ---" -ForegroundColor Yellow
$s3Buckets = aws s3 ls 2>$null | Select-String -Pattern "craftsbite|trainee"
Write-Host $s3Buckets
Write-Host ""

# 3. SSM Parameters (filter for craftsbite)
Write-Host "--- SSM Parameters (craftsbite) ---" -ForegroundColor Yellow
$ssmParams = aws ssm describe-parameters --region $Region `
    --parameter-type-filters "Type=SecureString" `
    --query "Parameters[?Name.startsWith('/craftsbite')]" `
    --output json 2>$null
if ($ssmParams) {
    $ssmParams | ConvertFrom-Json | ForEach-Object { $_.Name } | Write-Host
} else {
    # Fallback if query doesn't work
    aws ssm describe-parameters --region $Region --output text 2>$null | Select-String "craftsbite" | Write-Host
}
Write-Host ""

# 4. Lambda Functions (filter for craftsbite-related)
Write-Host "--- Lambda Functions ---" -ForegroundColor Yellow
$lambdaFunctions = aws lambda list-functions --region $Region --output json 2>$null
if ($lambdaFunctions) {
    $lambdaFunctions | ConvertFrom-Json | Select-Object -ExpandProperty Functions |
        Where-Object { $_.FunctionName -match "craftsbite|router|self|management|ops" } |
        Select-Object FunctionName, Runtime, Architecture, MemorySize, Timeout |
        Format-Table -AutoSize | Out-String | Write-Host
} else {
    aws lambda list-functions --region $Region --output text 2>$null | Write-Host
}
Write-Host ""

# 5. API Gateway (HTTP APIs)
Write-Host "--- API Gateway (HTTP APIs) ---" -ForegroundColor Yellow
$apis = aws apigatewayv2 get-apis --region $Region --output json 2>$null
if ($apis) {
    $apis | ConvertFrom-Json | Select-Object -ExpandProperty Items |
        Select-Object Name, ApiId, ApiEndpoint, ProtocolType |
        Format-Table -AutoSize | Out-String | Write-Host
} else {
    aws apigatewayv2 get-apis --region $Region --output text 2>$null | Write-Host
}
Write-Host ""

# 6. IAM Roles (filter for lambda-related)
Write-Host "--- IAM Roles (Lambda-related) ---" -ForegroundColor Yellow
$iamRoles = aws iam list-roles --output json 2>$null
if ($iamRoles) {
    $iamRoles | ConvertFrom-Json | Select-Object -ExpandProperty Roles |
        Where-Object { $_.RoleName -match "craftsbite|lambda|CraftsBite" } |
        Select-Object RoleName, Arn, CreatedDate |
        Format-Table -AutoSize | Out-String | Write-Host
}
Write-Host ""

# 7. CloudWatch Log Groups (Lambda-related)
Write-Host "--- CloudWatch Log Groups (Lambda-related) ---" -ForegroundColor Yellow
$logGroups = aws logs describe-log-groups --region $Region --output json 2>$null
if ($logGroups) {
    $logGroups | ConvertFrom-Json | Select-Object -ExpandProperty logGroups |
        Where-Object { $_.logGroupName -match "/aws/lambda/(craftsbite|router|self|management|ops)" } |
        Select-Object logGroupName, retentionInDays |
        Format-Table -AutoSize | Out-String | Write-Host
}
Write-Host ""

Write-Host "=== Discovery Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Green
Write-Host "1. Review the output above and note the exact resource names/IDs"
Write-Host "2. Update terraform/environments/prod.tfvars to match existing resource names"
Write-Host "3. Run .\scripts\import-existing.ps1 to import resources into Terraform state"