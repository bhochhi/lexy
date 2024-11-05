#!/bin/bash

# Exit on any error
set -e

# Set variables
ENV=${1:-dev}  # Use 'dev' as default if no argument is provided
AWS_PROFILE=${2:-dinospecs}  # Use 'dinospecs' as default if no argument is provided
TF_VAR_file="environments/${ENV}.tfvars"

# Get the current git branch name
BRANCH_NAME=$(git rev-parse --abbrev-ref HEAD)
# Remove any characters that are invalid in AWS resource names
BRANCH_SUFFIX=$(echo $BRANCH_NAME | sed 's/[^a-zA-Z0-9]/-/g')

# Check if we're in the right directory (should have terraform and lambda directories)
if [ ! -d "terraform" ]; then
    echo "Error: Script must be run from the project root directory"
    exit 1
fi

# Check if the environment file exists
if [ ! -f "$TF_VAR_file" ]; then
    echo "Error: Environment file $TF_VAR_file does not exist."
    exit 1
fi

# Export AWS_PROFILE
export AWS_PROFILE=$AWS_PROFILE

# Navigate to the Terraform directory
cd terraform

# Initialize Terraform
echo "Initializing Terraform..."
terraform init -upgrade

# Plan the changes
echo "Planning changes..."
terraform plan -var-file="../$TF_VAR_file" -var="branch_suffix=$BRANCH_SUFFIX" -out=tfplan

# Apply the changes
echo "Applying changes..."
terraform apply tfplan

# Clean up the plan file
echo "Cleaning up..."
rm tfplan

echo "✅ Deployment to $ENV environment completed"
echo "  Environment: $ENV"
echo "  AWS Profile: $AWS_PROFILE"
echo "  Branch: $BRANCH_NAME"
