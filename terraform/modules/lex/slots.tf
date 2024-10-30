resource "aws_lexv2models_slot" "slots" {
  for_each = { for slot in local.slots_map : slot.intent_name => slot }

  name         = each.value.name
  description  = each.value.description
  slot_type_id = each.value.slot_type_id
  bot_id       = aws_lexv2models_bot.chatbot.id
  bot_version  = var.lex_bot_version
  locale_id    = aws_lexv2models_bot_locale.bot_locale.locale_id
  intent_id    = split(":",aws_lexv2models_intent.intents[each.key].id)[0]

  value_elicitation_setting {
    default_value_specification {
      default_value_list {
        default_value = "Bank"
      }
    }
    slot_constraint = "Required"
    prompt_specification {
      allow_interrupt            = true
      max_retries                = 1
      message_selection_strategy = "Random"

      message_group {
        message {
          plain_text_message {
            value = each.value.prompt
          }
        }
      }

      prompt_attempts_specification {
        allow_interrupt = true
        map_block_key   = "Initial"

        allowed_input_types {
          allow_audio_input = true
          allow_dtmf_input  = true
        }

        audio_and_dtmf_input_specification {
          start_timeout_ms = 4000

          audio_specification {
            end_timeout_ms = 640
            max_length_ms  = 15000
          }

          dtmf_specification {
            deletion_character = "*"
            end_character      = "#"
            end_timeout_ms     = 5000
            max_length         = 513
          }
        }

        text_input_specification {
          start_timeout_ms = 30000
        }
      }

      prompt_attempts_specification {
        allow_interrupt = true
        map_block_key   = "Retry1"

        allowed_input_types {
          allow_audio_input = true
          allow_dtmf_input  = true
        }

        audio_and_dtmf_input_specification {
          start_timeout_ms = 4000

          audio_specification {
            end_timeout_ms = 640
            max_length_ms  = 15000
          }

          dtmf_specification {
            deletion_character = "*"
            end_character      = "#"
            end_timeout_ms     = 5000
            max_length         = 513
          }
        }

        text_input_specification {
          start_timeout_ms = 30000
        }
      }
    }
  }

  depends_on = [
    aws_lexv2models_intent.intents
  ]

}

