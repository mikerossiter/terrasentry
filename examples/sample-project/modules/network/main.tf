# Network module: portable networking primitives plus a managed Kubernetes
# cluster (EKS), which keeps workloads cloud-portable.

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_subnet" "private" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}

resource "aws_eks_cluster" "cluster" {
  name     = "acme-cluster"
  role_arn = "arn:aws:iam::123456789012:role/eks-cluster"
  vpc_config {
    subnet_ids = [aws_subnet.private.id]
  }
}
