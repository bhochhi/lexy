
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

variable "lex_bot_alias" {
  description = "The alias of the Lex bot"
  type        = string
}

variable "lambda_arn" {
  description = "The ARN of the Lambda function for Lex"
  type        = string
}

variable "lambda_function_name" {
  description = "The Lambda function for Lex"
  type        = string
}
