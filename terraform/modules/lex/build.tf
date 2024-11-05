# First build the bot

//create resource to run the build command for lex model so that you can send the user input to lex model
resource "null_resource" "build_lex_model" {
  triggers = {
    always_run = timestamp()
  }

  provisioner "local-exec" {
    command = "${path.module}/../../../scripts/build_model.sh ${aws_lexv2models_bot.chatbot.id} ${aws_lexv2models_bot_locale.bot_locale.locale_id}"
  }
  depends_on = [
    aws_lexv2models_bot_locale.bot_locale,
    aws_lexv2models_intent.welcome
  ]
}

# Outputs for monitoring
# output "build_status" {
#   value = fileexists("build_errors.txt") ? file("build_errors.txt") : "Build successful"
# }
#
# output "test_results" {
#   value = fileexists("test_results.json") ? jsondecode(file("test_results.json")) : "No test results available"
# }