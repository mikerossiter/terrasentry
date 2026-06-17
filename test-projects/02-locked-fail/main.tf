# Maximally cloud-locked AWS serverless stack: proprietary auth, NoSQL, FaaS,
# API gateway, step functions, and streaming. Almost impossible to move off
# AWS without a rewrite. Expected Terrasentry result: FAIL hard (repo ~0.14).

provider "aws" {
  region = "eu-west-2"
}

resource "aws_cognito_user_pool" "users" {
  name = "acme-users"
}

resource "aws_dynamodb_table" "sessions" {
  name         = "sessions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"

  attribute {
    name = "id"
    type = "S"
  }
}

resource "aws_lambda_function" "api" {
  function_name = "acme-api"
  runtime       = "nodejs20.x"
  handler       = "index.handler"
  role          = "arn:aws:iam::123456789012:role/lambda"
  filename      = "api.zip"
}

resource "aws_api_gateway_rest_api" "public" {
  name = "acme-public"
}

resource "aws_sfn_state_machine" "workflow" {
  name     = "acme-workflow"
  role_arn = "arn:aws:iam::123456789012:role/sfn"
  definition = jsonencode({ StartAt = "Done", States = { Done = { Type = "Succeed" } } })
}

resource "aws_kinesis_stream" "events" {
  name        = "acme-events"
  shard_count = 1
}
