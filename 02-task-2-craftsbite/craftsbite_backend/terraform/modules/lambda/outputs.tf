output "router_function_name" {
  description = "Router Lambda function name"
  value       = aws_lambda_function.router.function_name
}

output "router_arn" {
  description = "Router Lambda function ARN"
  value       = aws_lambda_function.router.arn
}

output "router_invoke_arn" {
  description = "Router Lambda invoke ARN (for API Gateway)"
  value       = aws_lambda_function.router.invoke_arn
}

output "gchat_router_function_name" {
  description = "Google Chat router Lambda function name"
  value       = aws_lambda_function.gchat_router.function_name
}

output "gchat_router_arn" {
  description = "Google Chat router Lambda function ARN"
  value       = aws_lambda_function.gchat_router.arn
}

output "gchat_router_invoke_arn" {
  description = "Google Chat router Lambda invoke ARN (for API Gateway)"
  value       = aws_lambda_function.gchat_router.invoke_arn
}

output "self_function_name" {
  description = "Self Lambda function name"
  value       = aws_lambda_function.self.function_name
}

output "self_arn" {
  description = "Self Lambda function ARN"
  value       = aws_lambda_function.self.arn
}

output "management_function_name" {
  description = "Management Lambda function name"
  value       = aws_lambda_function.management.function_name
}

output "management_arn" {
  description = "Management Lambda function ARN"
  value       = aws_lambda_function.management.arn
}

output "ops_function_name" {
  description = "Ops Lambda function name"
  value       = aws_lambda_function.ops.function_name
}

output "ops_arn" {
  description = "Ops Lambda function ARN"
  value       = aws_lambda_function.ops.arn
}

output "lambda_execution_role_arn" {
  description = "ARN of the shared Lambda execution IAM role"
  value       = aws_iam_role.lambda_execution.arn
}