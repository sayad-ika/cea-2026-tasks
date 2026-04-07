locals {
  s3_bucket = "trainee-2026-sayad-craftsbite"
  s3_prefix = "lambdas"

  lambdas = {
    "gchat-router" = {
      cmd_path        = "./cmd/gchat-router"
      lambda_function = "trainee-2026-sayad-craftsbite-gchat-router"
    }

    "management" = {
      cmd_path        = "./cmd/management"
      lambda_function = "trainee-2026-sayad-craftsbite-management"
    }

    "ops" = {
      cmd_path        = "./cmd/ops"
      lambda_function = "trainee-2026-sayad-craftsbite-ops"
    }

    "router" = {
      cmd_path        = "./cmd/router"
      lambda_function = "trainee-2026-sayad-craftsbite-router"
    }

    "self" = {
      cmd_path        = "./cmd/self"
      lambda_function = "trainee-2026-sayad-craftsbite-self"
    }
  }
}