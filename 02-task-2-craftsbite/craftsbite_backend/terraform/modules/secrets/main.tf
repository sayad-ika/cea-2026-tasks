resource "aws_ssm_parameter" "discord_bot_token" {
  name        = "/craftsbite/DISCORD_BOT_TOKEN"
  description = "Discord bot token"
  type        = "SecureString"
  value       = var.discord_bot_token

  tags = {
    Environment = var.environment
  }
}

resource "aws_ssm_parameter" "discord_public_key" {
  name        = "/craftsbite/DISCORD_PUBLIC_KEY"
  description = "Discord public key for signature verification"
  type        = "SecureString"
  value       = var.discord_public_key

  tags = {
    Environment = var.environment
  }
}

resource "aws_ssm_parameter" "gchat_service_account_json" {
  name        = "/craftsbite/gchat_service_account_json"
  description = "Google Chat service account JSON credentials"
  type        = "SecureString"
  value       = var.gchat_service_account_json

  tags = {
    Environment = var.environment
  }
}