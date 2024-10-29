
resource "aws_iam_role" "lex_bot_role" {
  name = "${var.project_name}-${var.environment}-lex-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lexv2.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })

  tags = {
    Project     = var.project_name
    Environment = var.environment
  }
}

resource "aws_iam_role_policy" "lex_bot_policy" {
  name = "${var.project_name}-${var.environment}-lex-policy"
  role = aws_iam_role.lex_bot_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lambda:InvokeFunction",
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "*"
      }
    ]
  })
}

# Lambda permission for Lex V2
resource "aws_lambda_permission" "lex_permission" {
  for_each      = { for idx, arn in tolist([var.lambda_arn]) : idx => arn }
  statement_id  = "AllowLexV2Invoke"
  action        = "lambda:InvokeFunction"
  function_name = split(":", each.value)[6]  # Extract function name from ARN
  principal     = "lexv2.amazonaws.com"
}
