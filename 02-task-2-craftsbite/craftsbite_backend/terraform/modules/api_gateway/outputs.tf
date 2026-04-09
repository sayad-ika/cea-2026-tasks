output "api_endpoint" {
  description = "API Gateway HTTP API endpoint URL"
  value       = aws_apigatewayv2_api.craftsbite.api_endpoint
}

output "api_execution_arn" {
  description = "API Gateway execution ARN"
  value       = aws_apigatewayv2_api.craftsbite.execution_arn
}

output "api_id" {
  description = "API Gateway ID"
  value       = aws_apigatewayv2_api.craftsbite.id
}