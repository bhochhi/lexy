resource "aws_lexv2models_intent" "welcome" {
  name        = "Welcome"
  description = "Welcome intent"
  bot_id      = aws_lexv2models_bot.chatbot.id
  bot_version = "DRAFT"
  locale_id   = aws_lexv2models_bot_locale.bot_locale.locale_id

  sample_utterance {
    utterance = "Hello"
  }
  sample_utterance {
    utterance = "Hi"
  }
  sample_utterance {
    utterance = "Hey"
  }

  closing_setting {
    closing_response {
      message_group {
        message {
          plain_text_message {
            value = "Hi there I am a bot. How can I help you today?"
          }
        }
      }
    }
  }



  depends_on = [
    aws_lexv2models_bot.chatbot,
    aws_lexv2models_bot_version.bot_version,
    aws_lexv2models_bot_locale.bot_locale
  ]
}