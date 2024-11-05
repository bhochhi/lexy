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

# variable "lex_bot_alias" {
#   description = "The alias of the Lex bot"
#   type        = string
# }

# variable "lex_bot_version" {
#   description = "The version of the Lex bot"
#   type        = string
#   default     = "DRAFT"
# }
