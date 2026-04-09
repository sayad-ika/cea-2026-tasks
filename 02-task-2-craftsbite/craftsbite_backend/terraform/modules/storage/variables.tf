variable "bucket_name" {
  description = "S3 bucket name for Lambda deployment artifacts"
  type        = string
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}