locals {
  name_prefix = "${var.project_name}-${var.environment}"

  common_env_vars = {
    AWS_REGION                      = var.aws_region
    DYNAMODB_TABLE                  = var.dynamodb_table_name
    CUTOFF_TIME                     = var.cutoff_time
    TIMEZONE                        = var.timezone
    DISCORD_APPLICATION_ID          = var.discord_application_id
    LAMBDA_SELF_FUNCTION_NAME       = var.lambda_self_function_name
    LAMBDA_MANAGEMENT_FUNCTION_NAME = var.lambda_management_function_name
    LAMBDA_OPS_FUNCTION_NAME        = var.lambda_ops_function_name
  }

  gchat_router_env_vars = merge(local.common_env_vars, {
    GCHAT_AUDIENCE = var.gchat_audience
  })
}

# S3 objects for Lambda deployment packages

resource "aws_s3_object" "router_zip" {
  bucket       = var.deployment_bucket
  key          = "lambdas/router.zip"
  source       = var.router_zip_path
  source_hash  = filemd5(var.router_zip_path)
  content_type = "application/zip"
}

resource "aws_s3_object" "gchat_router_zip" {
  bucket       = var.deployment_bucket
  key          = "lambdas/gchat-router.zip"
  source       = var.gchat_router_zip_path
  source_hash  = filemd5(var.gchat_router_zip_path)
  content_type = "application/zip"
}

resource "aws_s3_object" "self_zip" {
  bucket       = var.deployment_bucket
  key          = "lambdas/self.zip"
  source       = var.self_zip_path
  source_hash  = filemd5(var.self_zip_path)
  content_type = "application/zip"
}

resource "aws_s3_object" "management_zip" {
  bucket       = var.deployment_bucket
  key          = "lambdas/management.zip"
  source       = var.management_zip_path
  source_hash  = filemd5(var.management_zip_path)
  content_type = "application/zip"
}

resource "aws_s3_object" "ops_zip" {
  bucket       = var.deployment_bucket
  key          = "lambdas/ops.zip"
  source       = var.ops_zip_path
  source_hash  = filemd5(var.ops_zip_path)
  content_type = "application/zip"
}

# IAM Role - shared across all Lambda functions

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda_execution" {
  name               = "${local.name_prefix}-lambda-execution"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json

  tags = {
    Environment = var.environment
  }
}

# CloudWatch Logs policy

data "aws_iam_policy_document" "lambda_logging" {
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_role_policy" "lambda_logging" {
  name   = "${local.name_prefix}-lambda-logging"
  role   = aws_iam_role.lambda_execution.id
  policy = data.aws_iam_policy_document.lambda_logging.json
}

# DynamoDB access policy

data "aws_iam_policy_document" "dynamodb_access" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:UpdateItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:BatchGetItem",
      "dynamodb:BatchWriteItem",
    ]
    resources = [
      var.dynamodb_table_arn,
      "${var.dynamodb_table_arn}/index/*",
    ]
  }
}

resource "aws_iam_role_policy" "dynamodb_access" {
  name   = "${local.name_prefix}-dynamodb-access"
  role   = aws_iam_role.lambda_execution.id
  policy = data.aws_iam_policy_document.dynamodb_access.json
}

# SSM Parameter Store access policy

data "aws_iam_policy_document" "ssm_access" {
  statement {
    effect = "Allow"
    actions = [
      "ssm:GetParameter",
      "ssm:GetParameters",
    ]
    resources = [
      "arn:aws:ssm:${var.aws_region}:*:parameter/craftsbite/*",
    ]
  }
}

resource "aws_iam_role_policy" "ssm_access" {
  name   = "${local.name_prefix}-ssm-access"
  role   = aws_iam_role.lambda_execution.id
  policy = data.aws_iam_policy_document.ssm_access.json
}

# Lambda invoke policy (for router -> command lambda dispatch)

data "aws_iam_policy_document" "lambda_invoke" {
  statement {
    effect = "Allow"
    actions = [
      "lambda:InvokeFunction",
    ]
    resources = [
      aws_lambda_function.self.arn,
      aws_lambda_function.management.arn,
      aws_lambda_function.ops.arn,
    ]
  }
}

resource "aws_iam_role_policy" "lambda_invoke" {
  name   = "${local.name_prefix}-lambda-invoke"
  role   = aws_iam_role.lambda_execution.id
  policy = data.aws_iam_policy_document.lambda_invoke.json
}

# CloudWatch Log Groups (one per function, with 30-day retention)

resource "aws_cloudwatch_log_group" "router" {
  name              = "/aws/lambda/${local.name_prefix}-router"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "gchat_router" {
  name              = "/aws/lambda/${local.name_prefix}-gchat-router"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "self" {
  name              = "/aws/lambda/${var.lambda_self_function_name}"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "management" {
  name              = "/aws/lambda/${var.lambda_management_function_name}"
  retention_in_days = 30
}

resource "aws_cloudwatch_log_group" "ops" {
  name              = "/aws/lambda/${var.lambda_ops_function_name}"
  retention_in_days = 30
}

# Lambda Functions

resource "aws_lambda_function" "router" {
  function_name = "${local.name_prefix}-router"
  role          = aws_iam_role.lambda_execution.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  s3_bucket        = aws_s3_object.router_zip.bucket
  s3_key           = aws_s3_object.router_zip.key
  source_code_hash = filebase64sha256(var.router_zip_path)

  memory_size = var.lambda_router_memory
  timeout     = var.lambda_router_timeout

  environment {
    variables = local.common_env_vars
  }

  depends_on = [
    aws_iam_role_policy.lambda_logging,
    aws_iam_role_policy.dynamodb_access,
    aws_iam_role_policy.ssm_access,
    aws_iam_role_policy.lambda_invoke,
    aws_cloudwatch_log_group.router,
  ]

  tags = {
    Name        = "${local.name_prefix}-router"
    Environment = var.environment
  }
}

resource "aws_lambda_function" "gchat_router" {
  function_name = "${local.name_prefix}-gchat-router"
  role          = aws_iam_role.lambda_execution.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  s3_bucket        = aws_s3_object.gchat_router_zip.bucket
  s3_key           = aws_s3_object.gchat_router_zip.key
  source_code_hash = filebase64sha256(var.gchat_router_zip_path)

  memory_size = var.lambda_router_memory
  timeout     = var.lambda_router_timeout

  environment {
    variables = local.gchat_router_env_vars
  }

  depends_on = [
    aws_iam_role_policy.lambda_logging,
    aws_iam_role_policy.dynamodb_access,
    aws_iam_role_policy.ssm_access,
    aws_iam_role_policy.lambda_invoke,
    aws_cloudwatch_log_group.gchat_router,
  ]

  tags = {
    Name        = "${local.name_prefix}-gchat-router"
    Environment = var.environment
  }
}

resource "aws_lambda_function" "self" {
  function_name = var.lambda_self_function_name
  role          = aws_iam_role.lambda_execution.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  s3_bucket        = aws_s3_object.self_zip.bucket
  s3_key           = aws_s3_object.self_zip.key
  source_code_hash = filebase64sha256(var.self_zip_path)

  memory_size = var.lambda_command_memory
  timeout     = var.lambda_command_timeout

  environment {
    variables = local.common_env_vars
  }

  depends_on = [
    aws_iam_role_policy.lambda_logging,
    aws_iam_role_policy.dynamodb_access,
    aws_iam_role_policy.ssm_access,
    aws_cloudwatch_log_group.self,
  ]

  tags = {
    Name        = var.lambda_self_function_name
    Environment = var.environment
  }
}

resource "aws_lambda_function" "management" {
  function_name = var.lambda_management_function_name
  role          = aws_iam_role.lambda_execution.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  s3_bucket        = aws_s3_object.management_zip.bucket
  s3_key           = aws_s3_object.management_zip.key
  source_code_hash = filebase64sha256(var.management_zip_path)

  memory_size = var.lambda_command_memory
  timeout     = var.lambda_command_timeout

  environment {
    variables = local.common_env_vars
  }

  depends_on = [
    aws_iam_role_policy.lambda_logging,
    aws_iam_role_policy.dynamodb_access,
    aws_iam_role_policy.ssm_access,
    aws_cloudwatch_log_group.management,
  ]

  tags = {
    Name        = var.lambda_management_function_name
    Environment = var.environment
  }
}

resource "aws_lambda_function" "ops" {
  function_name = var.lambda_ops_function_name
  role          = aws_iam_role.lambda_execution.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"
  architectures = ["arm64"]

  s3_bucket        = aws_s3_object.ops_zip.bucket
  s3_key           = aws_s3_object.ops_zip.key
  source_code_hash = filebase64sha256(var.ops_zip_path)

  memory_size = var.lambda_command_memory
  timeout     = var.lambda_command_timeout

  environment {
    variables = local.common_env_vars
  }

  depends_on = [
    aws_iam_role_policy.lambda_logging,
    aws_iam_role_policy.dynamodb_access,
    aws_iam_role_policy.ssm_access,
    aws_cloudwatch_log_group.ops,
  ]

  tags = {
    Name        = var.lambda_ops_function_name
    Environment = var.environment
  }
}