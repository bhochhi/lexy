
provider "aws" {
  region = var.aws_region
}

locals {
  bot_name = "${var.project_name}_${var.environment}_${var.branch_suffix}"
  bot_alias = "${local.bot_name}_latest"
}

module "iam" {
  source = "./modules/iam"
  project_name = var.project_name
  # Add any necessary variables
}

module "lex" {
  source = "./modules/lex"
  lex_bot_name = "${local.bot_name}"
  lex_bot_alias = "${local.bot_alias}"
  lambda_arn = module.lambdas.lex_main_handler_arn
  # lambda_arn    = module.lambdas.lambda_arn
  depends_on    = [module.lambdas]
}


module "lambdas" {
  source = "./modules/lambdas/lex_main_handler"
  function_name   = var.lambda_function_name
  memory_size     = var.lambda_memory_size
  timeout         = var.lambda_timeout
  environment     = var.environment
  # role            = var.lambda_role_arn
  lambda_role_arn = module.iam.lambda_role_arn
}
