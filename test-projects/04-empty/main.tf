# An empty plan: a provider and variables are declared, but no resources are
# created. Exercises the "nothing to score" edge. Expected Terrasentry result:
# PASS (repo score n/a — no managed resources, so the lock-in gate does not
# apply). Before the fix this wrongly reported 0.00 / FAIL.

provider "aws" {
  region = "eu-west-2"
}

variable "environment" {
  type    = string
  default = "dev"
}
