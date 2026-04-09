# CraftsBite Terraform Import Guide

## What `terraform import` Does

Terraform manages infrastructure through a **state file** (`terraform.tfstate`). This file maps
Terraform resource addresses (like `module.database.aws_dynamodb_table.craftsbite`) to real AWS
resource IDs (like `arn:aws:dynamodb:ap-southeast-1:123456789012:table/craftsbite`).

When you run `terraform apply` for the FIRST time on an account that already has resources:

1. Terraform has **no state file** — it doesn't know any resources exist
2. Terraform sees its config says "create a DynamoDB table named X"
3. It tries to CREATE that table → **fails** because the name is already taken
4. It does NOT destroy or touch the existing table — it just errors out

`terraform import` bridges this gap. It tells Terraform: "Hey, this resource address in your
config — it already exists in AWS, here's its ID. Record that mapping in the state file so you
don't try to create it again."

After import, `terraform plan` will show **only the differences** between the existing resource's
current configuration and what your Terraform config declares. This lets you see exactly what
would change before making any modifications.

## The Two Approaches

### Approach A: Import Existing Resources (Recommended for Database)

Import the stateful resources that contain data you can't lose (DynamoDB table, S3 bucket, SSM
parameters). Let Terraform create new versions of stateless resources (Lambdas, IAM, API Gateway).

**Pros:**
- Zero risk to your DynamoDB data
- SSM secrets are preserved (no need to re-enter them)
- Clean new IAM roles and Lambda configs (no legacy cruft)
- Existing Lambda functions keep running until you swap to the new ones

**Cons:**
- Brief period where old and new Lambda functions coexist
- Need to update API Gateway routes to point to new Lambda functions

### Approach B: Import Everything

Import every single resource — DynamoDB, S3, SSM, Lambdas, IAM roles, API Gateway.

**Pros:**
- No resource duplication at all
- Terraform takes ownership of exactly what's running

**Cons:**
- Much more complex — need exact IDs for every resource
- High chance of naming mismatches (e.g., existing Lambda is `craftsbite-router` but Terraform
  wants `craftsbite-dev-router`)
- If ANY import fails, state gets partial and messy
- Your existing Lambdas are x86_64 but Terraform config specifies arm64 — `terraform plan`
  will show a forced replacement (destroy + recreate) for architecture change

## Step-by-Step: Approach A (Recommended)

### Prerequisites

1. AWS CLI configured with credentials for `ap-southeast-1`
2. Terraform >= 1.5.0 installed
3. Run `terraform -chdir=terraform init -backend=false`

### Step 0: Discover Existing Resources

Run the discovery script first to find exact resource names and IDs:

```powershell
.\scripts\discover-existing.ps1
```

This queries your AWS account and prints the IDs of all existing resources that match CraftsBite.
Save the output — you'll need these IDs for the import commands.

### Step 1: Update `prod.tfvars` to Match Existing Resource Names

This is critical. If your existing DynamoDB table is named `trainee-2026-sayad-craftsbite`
(check the DYNAMODB_TABLE in your .env), then your `prod.tfvars` must use THE SAME name:

```hcl
dynamodb_table_name = "trainee-2026-sayad-craftsbite"   # MUST match existing table
deployment_bucket_name = "trainee-2026-sayad-craftsbite"  # MUST match existing bucket
```

If you use a different name, Terraform will try to CREATE a new table and the import won't find
the existing one.

### Step 2: Run `terraform plan` First (Expect Errors)

```powershell
cd terraform
terraform plan -var-file=../terraform/environments/prod.tfvars
```

This will fail because resources already exist. That's expected — we're about to import them.

### Step 3: Import Stateful Resources

```powershell
# DynamoDB table (CRITICAL — your data lives here)
terraform import module.database.aws_dynamodb_table.craftsbite "trainee-2026-sayad-craftsbite"

# S3 deployment bucket
terraform import module.storage.aws_s3_bucket.deployment "trainee-2026-sayad-craftsbite"

# SSM Parameters
terraform import module.secrets.aws_ssm_parameter.discord_bot_token "/craftsbite/DISCORD_BOT_TOKEN"
terraform import module.secrets.aws_ssm_parameter.discord_public_key "/craftsbite/DISCORD_PUBLIC_KEY"
terraform import module.secrets.aws_ssm_parameter.gchat_service_account_json "/craftsbite/gchat_service_account_json"
```

### Step 4: Import S3 Bucket Sub-Resources (If They Exist)

Check if the bucket has versioning, encryption, etc. already configured:

```powershell
# Check if versioning is enabled
aws s3api get-bucket-versioning --bucket trainee-2026-sayad-craftsbite

# Check if public access block exists
aws s3api get-bucket-public-access-block --bucket trainee-2026-sayad-craftsbite

# Check if encryption is configured
aws s3api get-bucket-encryption --bucket trainee-2026-sayad-craftsbite
```

If these features already exist on the bucket, import them too:

```powershell
terraform import module.storage.aws_s3_bucket_versioning.deployment "trainee-2026-sayad-craftsbite"
terraform import module.storage.aws_s3_bucket_public_access_block.deployment "trainee-2026-sayad-craftsbite"
terraform import module.storage.aws_s3_bucket_server_side_encryption_configuration.deployment "trainee-2026-sayad-craftsbite"
```

Only import these if the features actually exist. If they don't, Terraform will create them on
the next `apply`.

### Step 5: Run `terraform plan` Again

```powershell
terraform plan -var-file=environments/prod.tfvars
```

Now the plan should show only:
- **New resources** to create (Lambda functions, IAM roles, API Gateway, CloudWatch log groups)
- **Minor changes** to imported resources (e.g., adding tags, enabling encryption)

**Important:** Review the plan carefully. You should see:
- `+ create` for new resources ✅
- `~ update in-place` for minor config drift ✅
- **NO** `- destroy` on the DynamoDB table ✅
- **NO** forced replacements on anything you imported ✅

### Step 6: Apply

```powershell
terraform apply -var-file=environments/prod.tfvars
```

This creates all the new resources (Lambdas, IAM, API Gateway) and updates the imported ones
to match the Terraform config.

### Step 7: Point API Gateway to New Lambdas

After the apply, your new API Gateway will route to the new Lambda functions. Update your
Discord and Google Chat webhook URLs to use the new API Gateway endpoint.

### Step 8: Verify and Clean Up

1. Test that all endpoints work via the new API Gateway
2. Delete the old Lambda functions manually (they're no longer referenced)
3. Delete the old API Gateway manually (if it exists)

---

## Troubleshooting

### "Error importing: resource already exists in state"

If you accidentally import something twice or need to re-import:

```powershell
# Remove from state (does NOT delete the AWS resource):
terraform state rm module.database.aws_dynamodb_table.craftsbite

# Then re-import:
terraform import module.database.aws_dynamodb_table.craftsbite "trainee-2026-sayad-craftsbite"
```

### "Error: Cannot import non-existent remote object"

The resource doesn't exist in AWS with that name. Check the exact name/ID using:

```powershell
aws dynamodb list-tables --region ap-southeast-1
aws s3 ls
aws ssm describe-parameters --region ap-southeast-1
```

### "Error: resource configuration doesn't match"

The imported resource's config differs from what Terraform declares. Run `terraform plan`
to see the diff. Common fixes:
- Update your tfvars to match the existing resource's config
- Accept the diff if it's a harmless change (like adding tags)

---

## State Backend (After First Successful Apply)

Once everything is working, move state from local to S3:

1. Create an S3 bucket for state and a DynamoDB table for locking (can use your existing
   deployment bucket or create a dedicated one)
2. Uncomment the `backend "s3"` block in `terraform/backend.tf`
3. Run `terraform init` — it will prompt to migrate state to S3

This ensures your team shares the same state file and enables state locking.