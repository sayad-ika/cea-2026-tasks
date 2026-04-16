terraform {
  required_version = ">= 1.5"

  backend "s3" {
    bucket  = "trainee-2026-sayad-craftsbite"
    key     = "terraform/state/terraform.tfstate"
    region  = "ap-southeast-1"
    encrypt = true
  }

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    null = {
      source  = "hashicorp/null"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}
