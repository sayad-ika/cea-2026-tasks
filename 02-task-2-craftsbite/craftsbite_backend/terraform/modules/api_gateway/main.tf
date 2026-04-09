resource "aws_apigatewayv2_api" "craftsbite" {
  name          = "${var.project_name}-${var.environment}-api"
  protocol_type = "HTTP"

  tags = {
    Environment = var.environment
  }
}

resource "aws_apigatewayv2_integration" "router" {
  api_id           = aws_apigatewayv2_api.craftsbite.id
  integration_type = "AWS_PROXY"
  integration_uri  = var.router_invoke_arn
}

resource "aws_apigatewayv2_integration" "gchat_router" {
  api_id           = aws_apigatewayv2_api.craftsbite.id
  integration_type = "AWS_PROXY"
  integration_uri  = var.gchat_router_invoke_arn
}

resource "aws_apigatewayv2_route" "interactions" {
  api_id    = aws_apigatewayv2_api.craftsbite.id
  route_key = "POST /interactions"
  target    = "integrations/${aws_apigatewayv2_integration.router.id}"
}

resource "aws_apigatewayv2_route" "gchat" {
  api_id    = aws_apigatewayv2_api.craftsbite.id
  route_key = "POST /gchat"
  target    = "integrations/${aws_apigatewayv2_integration.gchat_router.id}"
}

resource "aws_apigatewayv2_stage" "default" {
  api_id      = aws_apigatewayv2_api.craftsbite.id
  name        = "$default"
  auto_deploy = true

  tags = {
    Environment = var.environment
  }
}

resource "aws_lambda_permission" "api_gateway_router" {
  statement_id  = "AllowAPIGatewayInvoke-router"
  action        = "lambda:InvokeFunction"
  function_name = var.router_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.craftsbite.execution_arn}/*/POST/interactions"
}

resource "aws_lambda_permission" "api_gateway_gchat_router" {
  statement_id  = "AllowAPIGatewayInvoke-gchat-router"
  action        = "lambda:InvokeFunction"
  function_name = var.gchat_router_function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.craftsbite.execution_arn}/*/POST/gchat"
}