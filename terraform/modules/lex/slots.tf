resource "aws_lexv2models_slot" "slots" {
  for_each = {
    "Department" = {
      name         = "Department"
      description  = "Department slot"
      slot_type_id = "AMAZON.AlphaNumeric"
      intent_name  = "ContactUs"
      prompt       = "Which department would you like to contact?"
    }
  }

  name         = each.value.name
  description  = each.value.description
  slot_type_id = each.value.slot_type_id
  bot_id       = aws_lexv2models_bot.chatbot.id
  bot_version  = "DRAFT"
  locale_id    = aws_lexv2models_bot_locale.bot_locale.locale_id
  intent_id    = split(":", aws_lexv2models_intent.intents[each.value.intent_name].id)[0]
  value_elicitation_setting {
    slot_constraint = "Required"
    prompt_specification {
      message_selection_strategy = "Random"
      prompt_attempts_specification {
        map_block_key = "PromptAttemptsSpecification"
      }
      allow_interrupt = false
      max_retries     = 3
      message_group {
        message {
          plain_text_message {
            value = each.value.prompt
          }
        }
      }
    }
  }
}

# resource "aws_lexv2models_slot" "slots" {
#   for_each = { for slot in local.slots_map : slot.name => slot }
#
#   name         = each.value.name
#   description  = each.value.description
#   slot_type_id = each.value.slot_type_id
#   bot_id       = aws_lexv2models_bot.chatbot.id
#   bot_version  = "DRAFT"
#   locale_id    = aws_lexv2models_bot_locale.bot_locale.locale_id
#   intent_id    = substr(aws_lexv2models_intent.intents[each.value.intent_name].id, 0, 10)
#   value_elicitation_setting {
#     slot_constraint = "Required"
#     prompt_specification {
#       allow_interrupt            = true
#       max_retries                = 3
#       message_selection_strategy = "Random"
#       message_group {
#         message {
#           plain_text_message {
#             value = each.value.prompt
#           }
#         }
#       }
#     }
#   }
# }