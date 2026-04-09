# CraftsBite Terraform Infrastructure

Manages all AWS resources for the CraftsBite backend.

## Architecture

```
API Gateway (HTTP API)
  ├── POST /interactions → craftsbite-{env}-router Lambda
  └── POST /gchat        → craftsbite-{env}-gchat-router Lambda
          │
          ├── async invoke → craftsbite-{env}-self Lambda
          ├── async invoke → craftsbite-{env}-management Lambda
          └── async invoke → craftsbite-{env}-ops Lambda

DynamoDB Table: craftsbite(-{env}) [on-demand, PK+SK, GSI1]
S3 Bucket: trainee-2026-sayad-craftsbite(-{env}) [deployment artifacts]
SSM Parameters: /craftsbite/* [secrets - Discord token, public key, GChat SA]
IAM Role: Shared execution role with DynamoDB, SSM, Lambda invoke, CloudWatch
```

## Modules

| Module | Resources |
|--------|-----------|
| `storage` | S3 bucket, versioning, encryption, lifecycle, public access block |
| `database` | DynamoDB table (on-demand, PK+SK, GSI1), PITR |
| `secrets` | SSM SecureString parameters for Discord & GChat secrets |
| `lambda` | 5 Lambda functions (ARM64), IAM role + policies, CloudWatch log groups, S3 upload |
| `api_gateway` | HTTP API, 2 routes, 2 integrations, $default stage, Lambda permissions |
| `networking` | Placeholder for future VPC/config |

## Prerequisites

- Terraform >= 1.5.0
- AWS CLI configured with credentials for `ap-southeast-1`
- Go >= 1.22 (for building Lambdas)

## Quick Start (Fresh Deploy)

### 1. Create tfvars file

```bash
cp terraform/environments/dev.tfvars.example terraform/environments/dev.tfvars
```

Edit `dev.tfvars` — at minimum, set the required values:

```hcl
discord_application_id = "YOUR_DISCORD_APP_ID"

# Sensible defaults are already set for:
# aws_region, dynamodb_table_name, cutoff_time, timezone
```

Sensitive values (Discord bot token, Discord public key, GChat service account JSON) can be passed in two ways:

**Option A: Via environment variables** (recommended for CI/CD):
```bash
export TF_VAR_discord_bot_token="..."
export TF_VAR_discord_public_key="..."
export TF_VAR_gchat_service_account_json="..."
```

**Option B: Via tfvars file** (not recommended for version control):
Add them directly to `dev.tfvars`. Make sure `.tfvars` is in `.gitignore`.

### 2. Build Lambda binaries

```powershell
# Windows
.\scripts\build.ps1

# Linux/macOS
./scripts/build.sh
```

This compiles all 5 Lambdas for `GOOS=linux GOARCH=arm64` and creates zip files in `dist/`.

### 3. Initialize Terraform

```bash
terraform -chdir=terraform init
```

### 4. Plan and Apply

```bash
terraform -chdir=terraform plan -var-file=environments/dev.tfvars
terraform -chdir=terraform apply -var-file=environments/dev.tfvars
```

### 5. Two-Pass Deploy for GCHAT_AUDIENCE

The `GCHAT_AUDIENCE` env var on the gchat-router Lambda requires the API Gateway URL, which is
only available **after** the first `terraform apply`. This is a chicken-and-egg situation:

1. **First apply**: Leave `gchat_audience` empty (or set a placeholder). Note the `api_endpoint` output.
2. **Set the audience**: Update `dev.tfvars` with `gchat_audience = "https://abc123.execute-api.ap-southeast-1.amazonaws.com"`.
3. **Second apply**: `terraform apply` will update the gchat-router Lambda's `GCHAT_AUDIENCE` env var.

This only affects Google Chat integration. Discord works immediately after the first apply.

## Importing Existing Resources

If you have existing AWS resources (DynamoDB table, S3 bucket, SSM parameters) that were created
manually, you must import them into Terraform state before running `terraform apply`.

See [IMPORT-GUIDE.md](./IMPORT-GUIDE.md) for detailed instructions.

Quick steps:

```bash
# 1. Discover existing resources
.\scripts\discover-existing.ps1

# 2. Import stateful resources (safe — only updates state file)
.\scripts\import-existing.ps1

# 3. Review plan before applying
terraform -chdir=terraform plan -var-file=environments/dev.tfvars
```

## SSM Secrets

The following secrets are stored in AWS SSM Parameter Store as `SecureString`:

| Parameter | Description | Managed by |
|-----------|-------------|------------|
| `/craftsbite/DISCORD_BOT_TOKEN` | Discord bot OAuth token | Terraform |
| `/craftsbite/DISCORD_PUBLIC_KEY` | Discord Ed25519 public key | Terraform |
| `/craftsbite/gchat_service_account_json` | Google Chat service account JSON | Terraform |

The Go application reads these via `ssm:GetParameter` with decryption at Lambda cold start
(see `internal/config/config.go`). They are **not** passed as Lambda environment variables.

## Lambda Configuration

| Function | Memory | Timeout | Architecture | Runtime |
|----------|--------|---------|--------------|---------|
| router | 128 MB | 6s | arm64 | provided.al2 |
| gchat-router | 128 MB | 6s | arm64 | provided.al2 |
| self | 256 MB | 30s | arm64 | provided.al2 |
| management | 256 MB | 30s | arm64 | provided.al2 |
| ops | 256 MB | 30s | arm64 | provided.al2 |

Memory and timeout can be customized via tfvars:

```hcl
lambda_router_memory   = 256  # default: 128
lambda_router_timeout  = 10   # default: 6
lambda_command_memory  = 512  # default: 256
lambda_command_timeout = 60   # default: 30
```

## Environment Variables per Lambda

**All Lambdas receive:**
- `AWS_REGION`, `DYNAMODB_TABLE`, `CUTOFF_TIME`, `TIMEZONE`
- `DISCORD_APPLICATION_ID`
- `LAMBDA_SELF_FUNCTION_NAME`, `LAMBDA_MANAGEMENT_FUNCTION_NAME`, `LAMBDA_OPS_FUNCTION_NAME`

**gchat-router additionally receives:**
- `GCHAT_AUDIENCE` (the API Gateway invoke URL)

**Secrets fetched at runtime from SSM (not env vars):**
- `/craftsbite/DISCORD_BOT_TOKEN`
- `/craftsbite/DISCORD_PUBLIC_KEY`
- `/craftsbite/gchat_service_account_json`

## State Management

For local development, Terraform uses local state by default (no S3 backend needed).

For production/team use, uncomment the S3 backend block in `backend.tf` and run:

```bash
terraform -chdir=terraform init  # migrates state to S3
```

## Destroying Resources

```bash
terraform -chdir=terraform destroy -var-file=environments/dev.tfvars
```

**WARNING:** This permanently deletes all resources including the DynamoDB table and all data.
SSM parameter values are also destroyed. S3 bucket objects must be manually emptied first.