variable "project_name" {
  description = "The name of the project"
  type        = string
  default     = "lexy-chatbot"
}
variable "environment" {
  description = "Deployment environment"
  type        = string
}

variable "lex_bot_name" {
  description = "The name of the Lex bot"
  type        = string
}

variable "aws_region" {
  description = "The AWS region to deploy resources into"
  type        = string
  default     = "us-east-1"
}

# variable "lex_bot_alias" {
#   description = "The alias of the Lex bot"
#   type        = string
# }

# variable "lex_bot_version" {
#   description = "The version of the Lex bot"
#   type        = string
#   default     = "DRAFT"
# }
