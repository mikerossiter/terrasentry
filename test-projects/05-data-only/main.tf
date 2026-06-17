# A plan containing only data sources (lookups, not provisioning). Data sources
# are skipped by the scorer, so there are no managed resources to score.
# Exercises the data-source skip plus the "nothing to score" edge. Expected
# Terrasentry result: PASS (repo score n/a).

provider "aws" {
  region = "eu-west-2"
}

data "aws_caller_identity" "current" {}

data "aws_ami" "ubuntu" {
  most_recent = true
  owners      = ["099720109477"]
}
