
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
  lex_bot_name = local.bot_name
  # lex_bot_alias = local.bot_alias
  environment = var.environment
}

