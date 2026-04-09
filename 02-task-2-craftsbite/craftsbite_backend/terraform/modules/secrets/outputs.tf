output "discord_bot_token_arn" {
  description = "ARN of the Discord bot token SSM parameter"
  value       = aws_ssm_parameter.discord_bot_token.arn
}

output "discord_public_key_arn" {
  description = "ARN of the Discord public key SSM parameter"
  value       = aws_ssm_parameter.discord_public_key.arn
}

output "gchat_service_account_json_arn" {
  description = "ARN of the Google Chat service account JSON SSM parameter"
  value       = aws_ssm_parameter.gchat_service_account_json.arn
}