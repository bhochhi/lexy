
output "lex_bot_name" {
  description = "The name of the created Lex bot"
  value       = module.lex.bot_name
}

output "lambda_function_name" {
  description = "The name of the created Lambda function"
  value       = module.lambdas.function_name
}



output "debug_slots_by_intent" {
  description = "Debug output for slots by intent"
  value       = module.lex.debug_slots_by_intent
}

output "debug_slot_priorities" {
  description = "Debug output for slot priorities"
  value       = module.lex.debug_slot_priorities
}


output "all_intent_ids" {
  description = "Map of all intent IDs"
  value       =  module.lex.all_intent_ids
}
output "all_slot_ids" {
  value = module.lex.all_slot_ids
}

output "intents_debug" {
  description = "Debug output for intents"
  value       = module.lex.intents_debug
}

output "root_path" {
  value = path.root
}