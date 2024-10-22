terraform {
  backend "s3" {
    bucket = "lexy-chatbot-tf-state-bucket"
    key    = "lexy-chatbot/terraform.tfstate"
    region = "us-east-1"
  }
}