
resource "aws_lexv2models_intent" "greeting" {
  name        = "GreetingIntent"
  description = "Greeting Intent"
  locale_id   = var.locale_id
  bot_id      = var.bot_id
  bot_version = var.bot_version

  sample_utterance {
    utterance = "Hello"
  }

  sample_utterance {
    utterance = "Hi"
  }

  sample_utterance {
    utterance = "Hey there"
  }


  # Initial response when intent is recognized
  initial_response_setting {
    initial_response {
      message_group {
        message {
          plain_text_message {
            value = "Hello! How are you? How can I help you?"
          }
        }
      }
    }
  }

  closing_setting {
    closing_response {
      message_group {
        message {
          plain_text_message {
            value = "Hello! How are you? How can I help you?"
          }
        }
      }
    }
  }

  # Since we're using direct responses, we don't need a Lambda fulfillment
  fulfillment_code_hook {
    enabled = false
  }
}