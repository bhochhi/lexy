variable "function_name" {
  description = "The name of the Lambda function"
  type        = string
  default = "lex_main_handler"
}

variable "memory_size" {
  description = "The amount of memory for the Lambda function in MB"
  type        = number
  default     = 128
}

variable "timeout" {
  description = "The timeout for the Lambda function in seconds"
  type        = number
  default     = 30
}

variable "environment" {
  description = "The deployment environment (dev, stage, prod)"
  type        = string
  default     = "dev"
}
variable "lambda_role_arn" {
  description = "The ARN of the IAM role for the Lambda function"
  type        = string
}

# variable "lex_intent_arn" {
#   description = "ARN of the Lex intent"
#   type        = string
# }
