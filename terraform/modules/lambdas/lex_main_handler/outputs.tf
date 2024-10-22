output "function_name" {
  description = "The name of the created Lambda function"
  value       = aws_lambda_function.lex_handler.function_name
}

output "lex_main_handler_arn" {
  description = "The ARN of the created Lambda function"
  value       = aws_lambda_function.lex_handler.arn
}
output "lambda_arn" {
  description = "The ARN of the Lambda function"
  value       = aws_lambda_function.lex_handler.arn
}




