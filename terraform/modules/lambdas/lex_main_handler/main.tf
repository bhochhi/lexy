
locals {
  lambda_zip_path = "${path.module}/../../../../lambda/lex_main_handler/main.zip"
}

resource "aws_lambda_function" "lex_handler" {
  filename         = local.lambda_zip_path
  function_name    = var.function_name
  role            =  var.lambda_role_arn
  handler         = "main"
  runtime         = "provided.al2"
  source_code_hash = filebase64sha256(local.lambda_zip_path)

  environment {
    variables = {
      ENV = var.environment
    }
  }
}

# Add permission for Lex to invoke the Lambda function
resource "aws_lambda_permission" "lex_permission" {
  statement_id  = "AllowLexInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lex_handler.function_name
  principal     = "lex.amazonaws.com"

}
