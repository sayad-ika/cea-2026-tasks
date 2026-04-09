module "storage" {
  source      = "./modules/storage"
  bucket_name = var.deployment_bucket_name
  environment = var.environment
}

module "database" {
  source      = "./modules/database"
  table_name  = var.dynamodb_table_name
  environment = var.environment
}

module "secrets" {
  source                     = "./modules/secrets"
  discord_bot_token          = var.discord_bot_token
  discord_public_key         = var.discord_public_key
  gchat_service_account_json = var.gchat_service_account_json
  environment                = var.environment
}

module "lambda" {
  source = "./modules/lambda"

  environment  = var.environment
  project_name = var.project_name
  aws_region   = var.aws_region

  deployment_bucket = module.storage.bucket_id

  dynamodb_table_name = module.database.table_name
  dynamodb_table_arn  = module.database.table_arn

  discord_application_id = var.discord_application_id

  lambda_self_function_name       = "${var.project_name}-${var.environment}-self"
  lambda_management_function_name = "${var.project_name}-${var.environment}-management"
  lambda_ops_function_name        = "${var.project_name}-${var.environment}-ops"

  cutoff_time    = var.cutoff_time
  timezone       = var.timezone
  gchat_audience = var.gchat_audience

  lambda_router_memory   = var.lambda_router_memory
  lambda_router_timeout  = var.lambda_router_timeout
  lambda_command_memory  = var.lambda_command_memory
  lambda_command_timeout = var.lambda_command_timeout

  router_zip_path       = "${path.root}/../dist/router/router.zip"
  gchat_router_zip_path = "${path.root}/../dist/gchat-router/gchat-router.zip"
  self_zip_path         = "${path.root}/../dist/self/self.zip"
  management_zip_path   = "${path.root}/../dist/management/management.zip"
  ops_zip_path          = "${path.root}/../dist/ops/ops.zip"
}

module "api_gateway" {
  source = "./modules/api_gateway"

  project_name               = var.project_name
  environment                = var.environment
  router_invoke_arn          = module.lambda.router_invoke_arn
  gchat_router_invoke_arn    = module.lambda.gchat_router_invoke_arn
  router_function_name       = module.lambda.router_function_name
  gchat_router_function_name = module.lambda.gchat_router_function_name
}