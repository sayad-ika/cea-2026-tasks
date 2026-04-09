variable "project_name" {
  description = "Project name used as prefix"
  type        = string
  default     = "craftsbite"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}

variable "router_invoke_arn" {
  description = "Invoke ARN for the Discord router Lambda"
  type        = string
}

variable "gchat_router_invoke_arn" {
  description = "Invoke ARN for the Google Chat router Lambda"
  type        = string
}

variable "router_function_name" {
  description = "Function name for the Discord router Lambda"
  type        = string
}

variable "gchat_router_function_name" {
  description = "Function name for the Google Chat router Lambda"
  type        = string
}