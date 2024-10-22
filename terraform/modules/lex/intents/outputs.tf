
output "greeting_intent_id" {
  description = "ID of the greeting intent"
  value       = aws_lexv2models_intent.greeting.id
}

output "greeting_intent_name" {
  description = "Name of the greeting intent"
  value       = aws_lexv2models_intent.greeting.name
}

output "all_intent_ids" {
  description = "Map of all intent IDs"
  value = {
    greeting    = aws_lexv2models_intent.greeting.id
#     appointment = aws_lexv2models_intent.book_appointment.id
#     order       = aws_lexv2models_intent.check_order.id
  }
}