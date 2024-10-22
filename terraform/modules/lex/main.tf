resource "aws_lex_bot" "chatbot" {
  name = var.lex_bot_name
  description = "Lexy Chatbot"
  locale = "en-US"

  abort_statement {
    message {
      content = "Sorry, I couldn't understand. Goodbye."
      content_type = "PlainText"
    }
  }

  clarification_prompt {
    max_attempts = 2
    message {
      content = "I didn't understand you, could you please rephrase that?"
      content_type = "PlainText"
    }
  }

  idle_session_ttl_in_seconds = 300


  intent {
    intent_name    = aws_lex_intent.greeting.name
    intent_version = "$LATEST"
  }

   # Add process behavior for better reliability
  process_behavior = "BUILD"

  child_directed = false
}

resource "aws_lex_intent" "greeting" {
  name = "${var.lex_bot_name}_greeting"
  description = "Greeting intent for ${var.lex_bot_name}"

  sample_utterances = [
    "Hello",
    "Hi",
    "Hey there",
    "Greetings"
  ]

 fulfillment_activity {
    type = "CodeHook"
    code_hook {
      message_version = "1.0"
      uri             = var.lambda_arn
    }
  }


}



resource "aws_lex_bot_alias" "chatbot_alias" {
  bot_name = aws_lex_bot.chatbot.name
  name     = var.lex_bot_alias
  description = "Alias for ${var.lex_bot_name}"
  bot_version = "$LATEST"
}
