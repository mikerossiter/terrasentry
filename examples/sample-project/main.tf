# Sample Terraform project for testing Terrasentry. Deliberately mixes highly
# portable resources (VMs, object storage, Kubernetes) with cloud-locked ones
# (DynamoDB, Cognito, Lambda, API Gateway) so the scan produces an interesting
# lock-in score and several findings.
#
# To regenerate examples/sample-project/plan.json from this (requires AWS creds
# and `terraform init`):
#   terraform plan -out tfplan
#   terraform show -json tfplan > plan.json
#
# The committed plan.json is hand-maintained to match this file so the example
# runs with no cloud account.

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = "eu-west-2"
}

# --- Portable: commodity primitives -----------------------------------------

resource "aws_s3_bucket" "assets" {
  bucket = "acme-app-assets"
}

resource "aws_instance" "app" {
  ami           = "ami-0abcdef1234567890"
  instance_type = "t3.medium"
}

# --- Networking lives in a module --------------------------------------------

module "network" {
  source = "./modules/network"
}

# --- Cloud-locked: proprietary managed services ------------------------------

resource "aws_dynamodb_table" "sessions" {
  name         = "sessions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "id"
  attribute {
    name = "id"
    type = "S"
  }
}

resource "aws_cognito_user_pool" "users" {
  name = "acme-users"
}

resource "aws_lambda_function" "api" {
  function_name = "acme-api"
  runtime       = "nodejs20.x"
  handler       = "index.handler"
  filename      = "lambda.zip"
  role          = "arn:aws:iam::123456789012:role/lambda-exec"
}

resource "aws_api_gateway_rest_api" "public" {
  name = "acme-public-api"
}
