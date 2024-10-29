
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


output "all_intent_ids" {
  description = "Map of all intent IDs"
  value       = {for k, v in aws_lexv2models_intent.intents : k => v.id}
}

output "intents_debug" {
  description = "Debug output for intents"
  value       = local.intents
}

output "root_path" {
  value = path.root
}