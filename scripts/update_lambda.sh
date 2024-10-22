#!/bin/bash
# scripts/update_lambda.sh

set -e

# Ensure the AWS CLI is installed
if ! command -v aws &> /dev/null
then
    echo "AWS CLI could not be found. Please install it and try again."
    exit 1
fi

# Set variables
ENV=${1:-dev}  # Use 'dev' as default if no argument is provided
AWS_PROFILE=${2:-dinospecs}  # Use 'dinospecs' as default if no argument is provided

# Ensure the function name is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <lambda-function-name>"
    exit 1
fi

LAMBDA_FUNCTION_NAME=$1


# Check if AWS credentials are configured
if ! aws sts get-caller-identity &> /dev/null
then
    echo "AWS credentials not configured. Please run 'aws configure' and try again."
    exit 1
fi

# Disable AWS CLI pagination
export AWS_PAGER=""

# Build the Lambda function
./scripts/build_lambda.sh

# Navigate to the lambda function directory
cd lambda/lex_main_handler

# Update the Lambda function code
aws lambda update-function-code --function-name $LAMBDA_FUNCTION_NAME --zip-file fileb://main.zip --profile $AWS_PROFILE

echo "Lambda function code updated successfully."
