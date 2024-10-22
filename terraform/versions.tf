terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"  # Make sure to use version 5.x which supports LexV2
    }
  }
  required_version = ">= 1.0.0"
}

provider "aws" {
  region = var.aws_region
}