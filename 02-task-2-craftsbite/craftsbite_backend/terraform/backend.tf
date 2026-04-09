# Start with local state. Uncomment after running `terraform apply` once
# and the S3 bucket exists, then run `terraform init` to migrate state.
#
# terraform {
#   backend "s3" {
#     bucket         = "craftsbite-terraform-state"
#     key            = "craftsbite/terraform.tfstate"
#     region         = "ap-southeast-1"
#     dynamodb_table = "craftsbite-terraform-lock"
#     encrypt        = true
#   }
# }