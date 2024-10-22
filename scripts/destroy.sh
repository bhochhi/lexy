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

# Check if we're in the right directory
if [ ! -d "terraform" ] || [ ! -d "lambda" ]; then
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

# Print warning
echo "⚠️  WARNING! ⚠️"
echo "You are about to destroy all resources in:"
echo "  Environment: $ENV"
echo "  AWS Profile: $AWS_PROFILE"
echo "  Branch: $BRANCH_NAME"
echo ""

# Ask for confirmation
read -p "Are you sure you want to destroy these resources? (type 'yes' to confirm): " confirmation
if [ "$confirmation" != "yes" ]; then
    echo "Destruction cancelled."
    exit 0
fi

# Navigate to the Terraform directory
cd terraform

# Initialize Terraform (required to access state in S3)
echo "Initializing Terraform..."
terraform init -upgrade

# Plan the destruction
echo "Planning destruction..."
terraform plan -destroy -var-file="../$TF_VAR_file" -var="branch_suffix=$BRANCH_SUFFIX" -out=tfplan

# Final confirmation after showing plan
read -p "Review the plan above. Proceed with destruction? (type 'yes' to confirm): " final_confirmation
if [ "$final_confirmation" != "yes" ]; then
    echo "Destruction cancelled."
    rm tfplan
    exit 0
fi

# Apply the destruction
echo "Applying destruction..."
terraform apply tfplan

# Clean up the plan file
echo "Cleaning up..."
rm tfplan

echo "🗑️  Destruction completed"
echo "  Environment: $ENV"
echo "  AWS Profile: $AWS_PROFILE"
echo "  Branch: $
