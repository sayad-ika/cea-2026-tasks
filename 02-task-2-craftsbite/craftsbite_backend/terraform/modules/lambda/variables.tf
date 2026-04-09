variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "project_name" {
  description = "Project name used as prefix for resource naming"
  type        = string
  default     = "craftsbite"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "deployment_bucket" {
  description = "S3 bucket name for Lambda deployment artifacts"
  type        = string
}

variable "dynamodb_table_name" {
  description = "DynamoDB table name"
  type        = string
}

variable "dynamodb_table_arn" {
  description = "DynamoDB table ARN"
  type        = string
}

variable "discord_application_id" {
  description = "Discord application ID"
  type        = string
}

variable "lambda_self_function_name" {
  description = "Function name for the self Lambda"
  type        = string
}

variable "lambda_management_function_name" {
  description = "Function name for the management Lambda"
  type        = string
}

variable "lambda_ops_function_name" {
  description = "Function name for the ops Lambda"
  type        = string
}

variable "cutoff_time" {
  description = "Meal participation cutoff time (HH:MM)"
  type        = string
  default     = "21:00"
}

variable "timezone" {
  description = "Timezone for business logic"
  type        = string
  default     = "Asia/Dhaka"
}

variable "gchat_audience" {
  description = "Google Chat audience URL (API Gateway invoke URL)"
  type        = string
  default     = ""
}

variable "lambda_router_memory" {
  description = "Memory in MB for router Lambdas"
  type        = number
  default     = 128
}

variable "lambda_router_timeout" {
  description = "Timeout in seconds for router Lambdas"
  type        = number
  default     = 6
}

variable "lambda_command_memory" {
  description = "Memory in MB for command Lambdas"
  type        = number
  default     = 256
}

variable "lambda_command_timeout" {
  description = "Timeout in seconds for command Lambdas"
  type        = number
  default     = 30
}

# Zip file paths (absolute paths, computed from root module)

variable "router_zip_path" {
  description = "Absolute path to the router Lambda zip file"
  type        = string
}

variable "gchat_router_zip_path" {
  description = "Absolute path to the gchat-router Lambda zip file"
  type        = string
}

variable "self_zip_path" {
  description = "Absolute path to the self Lambda zip file"
  type        = string
}

variable "management_zip_path" {
  description = "Absolute path to the management Lambda zip file"
  type        = string
}

variable "ops_zip_path" {
  description = "Absolute path to the ops Lambda zip file"
  type        = string
}