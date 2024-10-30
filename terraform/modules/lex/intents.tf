resource "aws_lexv2models_intent" "intents" {
  for_each    = local.intents
  name        = each.value.name
  description = each.value.description
  bot_id      = aws_lexv2models_bot.chatbot.id
  bot_version = var.lex_bot_version
  locale_id   = aws_lexv2models_bot_locale.bot_locale.locale_id

  dynamic "sample_utterance" {
    for_each = each.value.sample_utterances
    content {
      utterance = sample_utterance.value
    }
  }


  # dynamic "slot_priority" {
  #   for_each = {for slot in local.slots_map : slot.name => slot if slot.intent_name == each.key}
  #   content {
  #     priority = slot_priority.value.priority
  #     slot_id  = aws_lexv2models_slot.slots[slot_priority.key].id
  #   }
  # }

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

  depends_on = [
    aws_lexv2models_bot.chatbot,
    aws_lexv2models_bot_locale.bot_locale
  ]
}


# wanted to just update slot priorities but it seems that its updates everything in the intent and does not preserve the existing values
# locals {
#   intent_ids = { for k, v in aws_lexv2models_intent.intents : k => split(":", v.id)[0] }
#   slot_ids   = { for k, v in aws_lexv2models_slot.slots : k => split(",", v.id)[0] }
# }
# resource "null_resource" "update_intent_slot_priority" {
#     triggers = {
#       always_run = timestamp()
#     }
#   for_each = { for slot in local.slots_map : slot.intent_name => slot }
#   provisioner "local-exec" {
#     command = <<EOT
#     aws lexv2-models update-intent \
#       --bot-id ${aws_lexv2models_bot.chatbot.id} \
#       --bot-version ${var.lex_bot_version} \
#       --locale-id ${aws_lexv2models_bot_locale.bot_locale.locale_id} \
#       --intent-id ${local.intent_ids[each.value.intent_name]} \
#       --intent-name ${each.value.intent_name} \
#       --slot-priorities "[{\"priority\": ${each.value.priority}, \"slotId\": \"${local.slot_ids[each.value.intent_name]}\"}]"
#     EOT
#   }
#   depends_on = [
#     aws_lexv2models_intent.intents,
#     aws_lexv2models_slot.slots
#   ]
# }

locals {
  intent_ids = { for k, v in aws_lexv2models_intent.intents : k => split(":", v.id)[0] }
  slot_ids   = { for k, v in aws_lexv2models_slot.slots : k => split(",", v.id)[2] }
  slots_by_intent = { for s in local.slots_map : s.intent_name => [for slot in local.slots_map : slot if slot.intent_name == s.intent_name] }

  slot_priorities = {
    for intent_name, slots in local.slots_by_intent :
    intent_name => [
      for slot in slots :
      {
        priority = slot.priority,
        slotId   = lookup(local.slot_ids, slot.name, null)
      } if lookup(local.slot_ids, slot.name, null) != null
    ]
  }
}

resource "null_resource" "update_intent_slot_priority" {
  triggers = {
    always_run = timestamp()
  }
  for_each = local.slot_priorities
  provisioner "local-exec" {
    command = <<EOT
      ${path.module}/../../../scripts/update_intent.sh \
      ${aws_lexv2models_bot.chatbot.id} \
      ${aws_lexv2models_bot_locale.bot_locale.locale_id} \
      ${var.lex_bot_version} \
      ${local.intent_ids[each.key]} \
      ${each.key} \
      '${jsonencode(each.value)}'
    EOT
  }
  depends_on = [
    aws_lexv2models_intent.intents,
    aws_lexv2models_slot.slots
  ]
}

output "debug_slots_by_intent" {
  value = local.slots_by_intent
}

output "debug_slot_priorities" {
  value = local.slot_priorities
}