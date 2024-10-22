
output "lex_bot_name" {
  description = "The name of the created Lex bot"
  value       = module.lex.lex_bot_name
}

output "lambda_function_name" {
  description = "The name of the created Lambda function"
  value       = module.lambdas.function_name
}
