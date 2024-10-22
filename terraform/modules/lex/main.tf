resource "aws_lexv2models_bot" "chatbot" {
  name        = var.lex_bot_name
  description = "Chatbot for ${var.project_name}"
  role_arn    = aws_iam_role.lex_bot_role.arn
  type        = "Bot"

  idle_session_ttl_in_seconds = 300

  data_privacy {
    child_directed = false
  }

  tags = {
    Project     = var.project_name
    Environment = var.environment
  }
}

resource "aws_lexv2models_bot_locale" "bot_locale" {
  bot_id      = aws_lexv2models_bot.chatbot.id
  bot_version = "DRAFT"
  locale_id   = "en_US"
  description = "English US locale for ${var.lex_bot_name}"

  n_lu_intent_confidence_threshold = 0.40
}

resource "aws_lexv2models_intent" "greeting" {
  name        = "GreetingIntent"
  description = "Greeting Intent"
  locale_id   = aws_lexv2models_bot_locale.bot_locale.locale_id
  bot_id      = aws_lexv2models_bot.chatbot.id
  bot_version = "DRAFT"

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

  depends_on = [aws_lexv2models_bot_locale.bot_locale]
}

# Create a bot version
resource "aws_lexv2models_bot_version" "bot_version" {
  bot_id = aws_lexv2models_bot.chatbot.id
  locale_specification = {
    "en_US" = {
        source_bot_version = "DRAFT"
    }
  }

  description = "Version for ${var.environment}"

  # Make sure both locale and intent are created before creating version
  depends_on = [
    aws_lexv2models_bot_locale.bot_locale,
    aws_lexv2models_intent.greeting
  ]
}

