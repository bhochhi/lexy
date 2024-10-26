resource "aws_lexv2models_intent" "intents" {
  for_each    = local.intents
  name        = each.value.name
  description = each.value.description
  bot_id      = aws_lexv2models_bot.chatbot.id
  bot_version = "DRAFT"
  # bot_version = aws_lexv2models_bot_version.bot_version.bot_id
  locale_id   = aws_lexv2models_bot_locale.bot_locale.locale_id

  dynamic "sample_utterance" {
    for_each = each.value.sample_utterances
    content {
      utterance = sample_utterance.value
    }
  }


  dynamic "fulfillment_code_hook" {
    for_each = lookup(each.value, "fulfillment_code_hook", null) != null ? [each.value.fulfillment_code_hook] : []
    content {
      enabled = fulfillment_code_hook.value.enabled
      post_fulfillment_status_specification {
        success_response {
          message_group {
            message {
              plain_text_message {
                value = fulfillment_code_hook.value.success_response
              }
            }
          }
        }
        failure_response {
          message_group {
            message {
              plain_text_message {
                value = fulfillment_code_hook.value.failure_response
              }
            }
          }
        }
      }
    }
  }

  dynamic "closing_setting" {
    for_each = lookup(each.value, "closing_responses", null) != null ? [each.value.closing_responses] : []
    content {
      active = true

      closing_response {
        message_group {
          message {
            plain_text_message {
              value = closing_setting.value.main_message
            }
          }
          dynamic "variation" {
            for_each = closing_setting.value.variations
            content {
              plain_text_message {
                value = variation.value
              }
            }
          }
        }
      }
    }
  }

  # dynamic "slot_priority" {
  #   for_each = each.value.slot_priorities
  #   content {
  #     slot_id = aws_lexv2models_slot.slots[each.key].id
  #     slot_name = slot_priority.value.slot_name
  #     priority  = slot_priority.value.priority
  #   }
  # }

  depends_on = [
    aws_lexv2models_bot.chatbot,
    aws_lexv2models_bot_version.bot_version,
    aws_lexv2models_bot_locale.bot_locale
  ]
}


//this is to break the cyclic dependency between slots and intents.
# resource "aws_lexv2models_intent" "intent_slot_priorities" {
#   for_each    = local.intents
#   name        = each.value.name
#   description = each.value.description
#   bot_id      = aws_lexv2models_bot.chatbot.id
#   bot_version = "DRAFT"
#   locale_id   = aws_lexv2models_bot_locale.bot_locale.locale_id
#
#   dynamic "slot_priority" {
#     for_each = each.value.slot_priorities
#     content {
#       slot_id   = aws_lexv2models_slot.slots[each.key].id
#       priority  = slot_priority.value.priority
#     }
#   }
#
#   depends_on = [
#     aws_lexv2models_slot.slots
#   ]
# }

output "all_intent_ids" {
  description = "Map of all intent IDs"
  value       = {for k, v in aws_lexv2models_intent.intents : k => v.id}
}

output "intents_debug" {
  description = "Debug output for intents"
  value       = local.intents
}