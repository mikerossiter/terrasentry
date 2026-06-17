# Portable-by-design AWS stack: generic compute, block + object storage,
# standard networking, and managed Kubernetes. Nothing here is hard to move to
# another cloud. Expected Terrasentry result: PASS (repo score ~0.78).

provider "aws" {
  region = "eu-west-2"
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "app" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_instance" "app" {
  ami           = "ami-0abcd1234efgh5678"
  instance_type = "t3.medium"
  subnet_id     = aws_subnet.app.id
}

resource "aws_ebs_volume" "data" {
  availability_zone = "eu-west-2a"
  size              = 100
}

resource "aws_s3_bucket" "assets" {
  bucket = "acme-portable-assets"
}

resource "aws_eks_cluster" "cluster" {
  name     = "acme-portable"
  role_arn = "arn:aws:iam::123456789012:role/eks"

  vpc_config {
    subnet_ids = [aws_subnet.app.id]
  }
}
