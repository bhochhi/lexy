output "lex_bot_name" {
  description = "The name of the created Lex bot"
  value       = aws_lex_bot.chatbot.name
}

output "intent_arn" {
  value = aws_lex_intent.greeting.arn
}
