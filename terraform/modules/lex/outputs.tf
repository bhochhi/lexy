
output "bot_id" {
  description = "The ID of the Lex bot"
  value       = aws_lexv2models_bot.chatbot.id
}

output "bot_name" {
  description = "The name of the Lex bot"
  value       = aws_lexv2models_bot.chatbot.name
}

output "bot_arn" {
  description = "The ARN of the Lex bot"
  value       = aws_lexv2models_bot.chatbot.arn
}

output "bot_version" {
  description = "The version of the bot"
  value       = aws_lexv2models_bot_version.bot_version.bot_version
}
