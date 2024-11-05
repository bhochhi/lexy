#!/bin/bash

# Exit on any error
set -e
cd terraform

# Validate the Terraform configuration
terraform validate

# Check if the validation was successful
if [ $? -eq 0 ]; then
  echo "Terraform configuration is valid."
else
  echo "Terraform configuration is invalid."
  exit 1
fi
cd ..