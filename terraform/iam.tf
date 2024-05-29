resource "aws_iam_role" "lex_execution_role" {
  name = "lex_execution_role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "lex_policy" {
  name   = "LexPolicy"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lex:PutIntent",
          "lex:CreateBotVersion",
          "lex:CreateIntentVersion",
          "lex:GetIntent"
        ],
        Resource = "*"
      },
      {
        Effect = "Allow",
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lex_policy_attachment" {
  role       = aws_iam_role.lex_execution_role.name
  policy_arn = aws_iam_policy.lex_policy.arn
}
