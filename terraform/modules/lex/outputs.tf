output "lex_bot_name" {
  description = "The name of the created Lex bot"
  value       = aws_lexv2models_bot.chatbot.name
}


output "intent_id" {
  description = "The ID of the intent"
  value       = aws_lexv2models_intent.greeting.id
}

output "intent_name" {
  description = "The name of the intent"
  value       = aws_lexv2models_intent.greeting.name
}

# Bot outputs
output "bot_id" {
  description = "The ID of the Lex bot"
  value       = aws_lexv2models_bot.chatbot.id
}

output "bot_arn" {
  description = "The ARN of the Lex bot"
  value       = aws_lexv2models_bot.chatbot.arn
}

# Version output
output "bot_version" {
  description = "The version of the bot"
  value       = aws_lexv2models_bot_version.bot_version.bot_version
}