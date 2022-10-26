terraform {
  required_version = ">= 1.3.0"

  required_providers {
    aws = {
      source = "hashicorp/aws"
    }
  }
}

provider "aws" {
  region = "eu-west-2"
}

variable "username" {
  type        = string
  description = "Basic authentication username"
  sensitive   = true
}

variable "password" {
  type        = string
  description = "Basic authentication password"
  sensitive   = true
}

locals {
  lambda_handler = "optimized-m3u-iptv-list-server"
  name_prefix    = "siptv-list-optimizer"
}

data "archive_file" "lambda_zip" {
  type        = "zip"
  source_dir  = "../bin"
  output_path = "bin/optimized-m3u-iptv-list-server.zip"
}

# IAM ##########################################

data "aws_iam_policy_document" "assume_role" {
  policy_id = "${local.name_prefix}-lambda"
  version   = "2012-10-17"
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name               = "${local.name_prefix}-lambda"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}


data "aws_iam_policy_document" "logs" {
  policy_id = "${local.name_prefix}-lambda-logs"
  version   = "2012-10-17"
  statement {
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }
}

resource "aws_iam_policy" "logs" {
  name   = "${local.name_prefix}-lambda-logs"
  policy = data.aws_iam_policy_document.logs.json
}

resource "aws_iam_role_policy_attachment" "logs" {
  depends_on = [aws_iam_role.lambda, aws_iam_policy.logs]
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.logs.arn
}

# CLOUDWATCH ##########################################

resource "aws_cloudwatch_log_group" "log" {
  name              = "/aws/lambda/${local.name_prefix}"
  retention_in_days = 7
}

# LAMBDA FUNCTION #####################################

resource "aws_lambda_function" "optimize_siptv_list_lambda" {
  filename         = data.archive_file.lambda_zip.output_path
  function_name    = "${local.name_prefix}-handler"
  role             = aws_iam_role.lambda.arn
  handler          = local.lambda_handler
  source_code_hash = filebase64sha256(data.archive_file.lambda_zip.output_path)
  runtime          = "go1.x"
  memory_size      = 1024
  timeout          = 60

  depends_on = [
    aws_iam_role_policy_attachment.logs,
    aws_cloudwatch_log_group.log,
  ]

  environment {
    variables = {
      USERNAME = var.username
      PASSWORD = var.password
    }
  }
}

resource "aws_lambda_permission" "apigw" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.optimize_siptv_list_lambda.arn
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.siptv_opmtizer.execution_arn}/*/*"
}


# API GW #####################################

resource "aws_apigatewayv2_api" "siptv_opmtizer" {
  name          = "${local.name_prefix}-api"
  protocol_type = "HTTP"
  description   = "This API optimizes a clutered SIPTV list."
  target        = aws_lambda_function.optimize_siptv_list_lambda.arn
}
