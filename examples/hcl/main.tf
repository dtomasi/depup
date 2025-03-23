# terraform/main.tf - Example Terraform configuration for depup

# Configure the AWS provider
# depup package=aws-provider
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "4.0.0"
    }

    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.0.0" // depup package=azure-provider
    }
  }
}

# Configure AWS provider settings
provider "aws" {
  region = "us-west-2"
}

# VPC module with version annotation
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  // depup package=aws-vpc-module
  version = "3.14.0"

  name = "my-vpc"
  cidr = "10.0.0.0/16"
  azs  = ["us-west-2a", "us-west-2b", "us-west-2c"]
}

# S3 bucket module
module "s3_bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "2.11.1" # depup package=aws-s3-module

  bucket = "my-example-bucket"
  acl    = "private"
}
