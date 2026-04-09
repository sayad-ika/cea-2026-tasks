output "api_endpoint" {
  description = "API Gateway HTTP API endpoint URL"
  value       = module.api_gateway.api_endpoint
}

output "dynamodb_table_name" {
  description = "DynamoDB table name"
  value       = module.database.table_name
}

output "deployment_bucket_name" {
  description = "S3 deployment bucket name"
  value       = module.storage.bucket_id
}

output "lambda_function_names" {
  description = "Lambda function names"
  value = {
    router       = module.lambda.router_function_name
    gchat_router = module.lambda.gchat_router_function_name
    self         = module.lambda.self_function_name
    management   = module.lambda.management_function_name
    ops          = module.lambda.ops_function_name
  }
}