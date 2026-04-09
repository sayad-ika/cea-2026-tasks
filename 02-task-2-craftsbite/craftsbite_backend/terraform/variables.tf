variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "ap-southeast-1"
}

variable "project_name" {
  description = "Project name used as prefix for all resources"
  type        = string
  default     = "craftsbite"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "dynamodb_table_name" {
  description = "DynamoDB table name"
  type        = string
  default     = "craftsbite"
}

variable "deployment_bucket_name" {
  description = "S3 bucket name for Lambda deployment artifacts"
  type        = string
  default     = "trainee-2026-sayad-craftsbite"
}

variable "cutoff_time" {
  description = "Meal participation cutoff time (HH:MM format)"
  type        = string
  default     = "21:00"
}

variable "timezone" {
  description = "Timezone for business logic"
  type        = string
  default     = "Asia/Dhaka"
}

variable "discord_application_id" {
  description = "Discord application ID"
  type        = string
  sensitive   = false
}

variable "discord_bot_token" {
  description = "Discord bot token (stored in SSM)"
  type        = string
  sensitive   = true
}

variable "discord_public_key" {
  description = "Discord public key (stored in SSM)"
  type        = string
  sensitive   = true
}

variable "gchat_service_account_json" {
  description = "Google Chat service account JSON (stored in SSM)"
  type        = string
  sensitive   = true
}

variable "gchat_audience" {
  description = "Google Chat audience URL (API Gateway invoke URL)"
  type        = string
  default     = ""
}

# Lambda configuration variables

variable "lambda_router_memory" {
  description = "Memory in MB for router Lambda"
  type        = number
  default     = 128
}

variable "lambda_router_timeout" {
  description = "Timeout in seconds for router Lambda"
  type        = number
  default     = 6
}

variable "lambda_command_memory" {
  description = "Memory in MB for command Lambdas (self, management, ops)"
  type        = number
  default     = 256
}

variable "lambda_command_timeout" {
  description = "Timeout in seconds for command Lambdas (self, management, ops)"
  type        = number
  default     = 30
}