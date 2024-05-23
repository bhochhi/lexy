provider "aws" {
  region = var.region
}

resource "aws_lex_bot" "my_lexy" {
  name        = "MyLexy"
  description = "A bot for handling financial service inquiries"
  intents     = [ for intent in local.intents : {
    intent_name       = intent.name
    intent_version    = "$LATEST"
  }]
}

resource "null_resource" "upload_intents" {
  provisioner "local-exec" {
    command = "go run ${path.module}/../scripts/upload_intents.go"
  }

  triggers = {
    intents_hash = filebase64sha256("${path.module}/../data/intents.json")
  }
}
