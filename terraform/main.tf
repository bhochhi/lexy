provider "aws" {
  region = var.region
}

resource "aws_lex_bot" "my_lexy" {
  name        = "MyLexyBot"
  description = "A bot for handling financial service inquiries"
  intents     = [ for intent in local.intents : {
    intent_name       = intent.name
    intent_version    = "$LATEST"
  }]
}

resource "null_resource" "upload_intents" {
  provisioner "local-exec" {
    command = "go run ${path.module}/../scripts/upload_intents.go"
    environment = {
      AWS_REGION = var.region
      AWS_ACCESS_KEY_ID = var.aws_access_key_id
      AWS_SECRET_ACCESS_KEY = var.aws_secret_access_key
    }
  }

  triggers = {
    intents_hash = filebase64sha256("${path.module}/../data/intents.json")
  }
}

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
        Effect = "Allow",
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
