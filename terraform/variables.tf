
variable "aws_region" {
  description = "The AWS region to deploy resources into"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "The name of the project"
  type        = string
  default     = "lexy-chatbot"
}

variable "lex_bot_name" {
  description = "The name of the Lex bot"
  type        = string
  default = "lexy_chatbot"
}

variable "lex_bot_alias" {
  description = "The alias of the Lex bot"
  type        = string
  default = "lexy_chatbot_latest"
}

variable "branch_suffix" {
  description = "Suffix to add to resource names based on the git branch"
  type        = string
  default     = ""
}

variable "environment" {
  description = "Deployment environment"
  type        = string
  default     = "dev"
}

variable "bot_purpose" {
  description = "Purpose of the bot"
  type        = string
  default     = "customer-support"
}

variable "lambda_function_name" {
  description = "The name of the Lambda function"
  type        = string
}

variable "lambda_memory_size" {
  description = "The amount of memory for the Lambda function in MB"
  type        = number
}

variable "lambda_timeout" {
  description = "The timeout for the Lambda function in seconds"
  type        = number
}
