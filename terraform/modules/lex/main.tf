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
    module.intents
  ]
}



# Call the intents module
module "intents" {
  source = "./intents"

  bot_id        = aws_lexv2models_bot.chatbot.id
  bot_version   = "DRAFT"
  locale_id     = aws_lexv2models_bot_locale.bot_locale.locale_id
}

