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