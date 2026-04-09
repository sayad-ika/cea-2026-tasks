variable "discord_bot_token" {
  description = "Discord bot token"
  type        = string
  sensitive   = true
}

variable "discord_public_key" {
  description = "Discord public key for signature verification"
  type        = string
  sensitive   = true
}

variable "gchat_service_account_json" {
  description = "Google Chat service account JSON credentials"
  type        = string
  sensitive   = true
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
}