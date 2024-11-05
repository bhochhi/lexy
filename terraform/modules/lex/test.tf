resource "null_resource" "test_bot" {
  triggers = {
    build_completed = null_resource.build_lex_model.id
  }

  provisioner "local-exec" {
    command     = "${path.module}/../../../scripts/run_tests.sh ${aws_lexv2models_bot.chatbot.id} ${var.aws_region}"
  }

  depends_on = [
    null_resource.build_lex_model
  ]
}