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
  bot_version = var.lex_bot_version
  locale_id   = "en_US"
  description = "English US locale for ${var.lex_bot_name}"

  n_lu_intent_confidence_threshold = 0.40
  depends_on = [aws_lexv2models_bot.chatbot]

}



//create resource to run the build command for lex model so that you can send the user input to lex model
resource "null_resource" "build_lex_model" {
  triggers = {
    always_run = timestamp()
  }

  provisioner "local-exec" {
    command = "${path.module}/../../../scripts/build_model.sh ${aws_lexv2models_bot.chatbot.id} ${aws_lexv2models_bot_locale.bot_locale.locale_id} ${var.lex_bot_version}"
  }
  depends_on = [
    aws_lexv2models_bot_locale.bot_locale,
    aws_lexv2models_intent.intents,
    aws_lexv2models_slot.slots
  ]
}