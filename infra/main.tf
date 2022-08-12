terraform {
  cloud {
    organization = "brennonloveless-personal"

    workspaces {
      name = "golang-graphql-stock-tracker"
    }
  }
}

provider "aws" {
  region = "us-west-2"
  profile = "personal"

  default_tags {
    tags = {
      Environment = "Production"
      Product = "Golang Graphql Stock Tracker"
    }
  }
}

variable "app_name" {
  description = "Application name"
  default     = "Golang Graphql Stock Tracker"
}

variable "app_env" {
  description = "Application environment tag"
  default     = "dev"
}

locals {
  app_id = "${lower(var.app_name)}-${lower(var.app_env)}-${random_id.unique_suffix.hex}"
}

data "archive_file" "graphql_zip" {
  type        = "zip"
  source_file = "../build/bin/graphql"
  output_path = "../build/bin/graphql.zip"
}

data "archive_file" "stock_poller_zip" {
  type        = "zip"
  source_file = "../build/bin/stockpoller"
  output_path = "../build/bin/stockpoller.zip"
}

resource "random_id" "unique_suffix" {
  byte_length = 4
}

output "api_url" {
  value = aws_apigatewayv2_api.lambda.api_endpoint
}

resource "aws_cloudwatch_log_group" "graphql" {
  name = "/aws/lambda/${aws_lambda_function.graphql.function_name}"

  retention_in_days = 30
}

resource "aws_iam_role_policy_attachment" "graphql_lambda_policy" {
  role       = aws_iam_role.graphql.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy_attachment" "stock_poller_lambda_policy" {
  role       = aws_iam_role.stock_poller.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_apigatewayv2_api" "lambda" {
  name          = "serverless_lambda_gw"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_stage" "lambda" {
  api_id = aws_apigatewayv2_api.lambda.id

  name        = "dev"
  auto_deploy = true

  access_log_settings {
    destination_arn = aws_cloudwatch_log_group.api_gw.arn

    format = jsonencode({
      requestId               = "$context.requestId"
      sourceIp                = "$context.identity.sourceIp"
      requestTime             = "$context.requestTime"
      protocol                = "$context.protocol"
      httpMethod              = "$context.httpMethod"
      resourcePath            = "$context.resourcePath"
      routeKey                = "$context.routeKey"
      status                  = "$context.status"
      responseLength          = "$context.responseLength"
      integrationErrorMessage = "$context.integrationErrorMessage"
    }
  )
}
}

resource "aws_apigatewayv2_integration" "graphql" {
  api_id = aws_apigatewayv2_api.lambda.id

  integration_uri    = aws_lambda_function.graphql.invoke_arn
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "graphql" {
  api_id = aws_apigatewayv2_api.lambda.id

  route_key = "POST /graphql"
  target    = "integrations/${aws_apigatewayv2_integration.graphql.id}"
}

resource "aws_cloudwatch_log_group" "api_gw" {
  name = "/aws/api_gw/${aws_apigatewayv2_api.lambda.name}"

  retention_in_days = 30
}

resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.graphql.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.lambda.execution_arn}/*/*"
}

resource "aws_lambda_function" "graphql" {
  filename         = data.archive_file.graphql_zip.output_path
  function_name    = "graphql-${lower(var.app_env)}-${random_id.unique_suffix.hex}"
  handler          = "graphql"
  source_code_hash = base64sha256(data.archive_file.graphql_zip.output_path)
  runtime          = "go1.x"
  role             = aws_iam_role.graphql.arn

  environment {
    variables = {
      USER_ORDERS_TABLE = aws_dynamodb_table.user_orders.id
    }
  }
}

resource "aws_lambda_function" "stock_poller" {
  filename         = data.archive_file.stock_poller_zip.output_path
  function_name    = "stock-poller-${lower(var.app_env)}-${random_id.unique_suffix.hex}"
  handler          = "stockpoller"
  source_code_hash = base64sha256(data.archive_file.stock_poller_zip.output_path)
  runtime          = "go1.x"
  role             = aws_iam_role.stock_poller.arn
  timeout          = 60

  environment {
    variables = {
      TRACKED_STOCKS_TABLE = aws_dynamodb_table.tracked_stocks.id
      STOCKS_TABLE = aws_dynamodb_table.stocks.id
    }
  }
}

resource "aws_cloudwatch_event_rule" "pollstocks_interval" {
  name = "pollstocks-interval"
  description = "Triggers the stock polling lambda function to gather a new set of data"
  schedule_expression = "rate(15 minutes)"
}

resource "aws_cloudwatch_event_target" "pollstocks_interval" {
  rule = aws_cloudwatch_event_rule.pollstocks_interval.name
  target_id = "stock_poller"
  arn = aws_lambda_function.stock_poller.arn
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_stock_poller" {
  statement_id = "AllowExecutionFromCloudWatch"
  action = "lambda:InvokeFunction"
  function_name = aws_lambda_function.stock_poller.function_name
  principal = "events.amazonaws.com"
  source_arn = aws_cloudwatch_event_rule.pollstocks_interval.arn
}

# Assume role setup
resource "aws_iam_role" "graphql" {
  name_prefix = "graphql-${lower(var.app_env)}-${random_id.unique_suffix.hex}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Effect = "Allow"
        Sid = ""
      }
    ]
  })

  inline_policy {
    name = "DynamoWriter"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "dynamodb:BatchGetItem",
            "dynamodb:BatchWriteItem",
            "dynamodb:GetItem",
            "dynamodb:PutItem",
            "dynamodb:Scan",
            "dynamodb:UpdateItem",
          ]
          Effect = "Allow"
          Resource = [aws_dynamodb_table.user_orders.arn]
        }
      ]
    })
  }
}

# Attach role to Managed Policy
variable "iam_policy_arn" {
  description = "IAM Policy to be attached to role"
  type        = list(string)

  default = [
    "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  ]
}

resource "aws_iam_policy_attachment" "graphql_role_attach" {
  name       = "graphql-policy-${local.app_id}"
  roles      = [aws_iam_role.graphql.id]
  count      = length(var.iam_policy_arn)
  policy_arn = element(var.iam_policy_arn, count.index)
}

# Assume role setup
resource "aws_iam_role" "stock_poller" {
  name_prefix = "stock-poller-${lower(var.app_env)}-${random_id.unique_suffix.hex}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Effect = "Allow"
        Sid = ""
      }
    ]
  })

  inline_policy {
    name = "SecretsReader"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "secretsmanager:DescribeSecret",
            "secretsmanager:GetSecretValue",
          ]
          Effect = "Allow"
          Resource = ["arn:aws:secretsmanager:us-west-2:391324319136:secret:prod/GraphQLStocks-mERzvd"]
        }
      ]
    })
  }

  inline_policy {
    name = "DynamoWriter"

    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "dynamodb:BatchGetItem",
            "dynamodb:BatchWriteItem",
            "dynamodb:GetItem",
            "dynamodb:PutItem",
            "dynamodb:Query",
            "dynamodb:UpdateItem",
            "dynamodb:DeleteItem",
          ]
          Effect = "Allow"
          Resource = [
            aws_dynamodb_table.stocks.arn,
            aws_dynamodb_table.tracked_stocks.arn,
          ]
        }
      ]
    })
  }
}

resource "aws_iam_policy_attachment" "stock_poller_role_attach" {
  name       = "stock-poller-policy-${local.app_id}"
  roles      = [aws_iam_role.stock_poller.id]
  count      = length(var.iam_policy_arn)
  policy_arn = element(var.iam_policy_arn, count.index)
}

resource "aws_dynamodb_table" "user_orders" {
  name = "UserOrders-${random_id.unique_suffix.hex}"
  billing_mode = "PROVISIONED"
  read_capacity = 5
  write_capacity = 5
  hash_key = "PK"

  attribute {
    name = "PK"
    type = "S"
  }

  ttl {
    attribute_name = "DeletedAt"
    enabled = true
  }
}

resource "aws_dynamodb_table" "stocks" {
  name = "Stocks-${random_id.unique_suffix.hex}"
  billing_mode = "PROVISIONED"
  read_capacity = 5
  write_capacity = 5
  hash_key = "PK"
  range_key = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  ttl {
    attribute_name = "deleted_at"
    enabled = true
  }
}

resource "aws_dynamodb_table" "tracked_stocks" {
  name = "TrackedStocks-${random_id.unique_suffix.hex}"
  billing_mode = "PROVISIONED"
  read_capacity = 5
  write_capacity = 5
  hash_key = "enabled"
  range_key = "last_polled"

  attribute {
    name = "enabled"
    type = "S"
  }

  attribute {
    name = "last_polled"
    type = "S"
  }

  ttl {
    attribute_name = "deleted_at"
    enabled = true
  }
}

