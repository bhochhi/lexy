variable "function_name" {
  description = "The name of the Lambda function"
  type        = string
}

variable "memory_size" {
  description = "The amount of memory for the Lambda function in MB"
  type        = number
}

variable "timeout" {
  description = "The timeout for the Lambda function in seconds"
  type        = number
}

variable "environment" {
  description = "The deployment environment (dev, stage, prod)"
  type        = string
}
variable "lambda_role_arn" {
  description = "The ARN of the IAM role for the Lambda function"
  type        = string
}

# variable "lex_intent_arn" {
#   description = "ARN of the Lex intent"
#   type        = string
# }
