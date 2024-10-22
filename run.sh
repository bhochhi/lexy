# #!/bin/bash

# # Create main Terraform directory and files
# mkdir -p terraform
# touch terraform/{main.tf,variables.tf,outputs.tf,versions.tf,backend.tf}

# # Create Terraform modules directory and Lambda handler module
# mkdir -p terraform/modules/lambdas/lex_main_handler
# touch terraform/modules/lambdas/lex_main_handler/{main.tf,variables.tf,outputs.tf}

# # Create Terraform modules for Lex and IAM
# mkdir -p terraform/modules/{lex,iam}
# touch terraform/modules/lex/{main.tf,variables.tf,outputs.tf}
# touch terraform/modules/iam/{main.tf,variables.tf,outputs.tf}

# # Create Lambda source code directory and main Go file
# mkdir -p lambda/lex_main_handler
# touch lambda/lex_main_handler/main.go

# # Create environments directory and tfvars files
# mkdir -p environments
# touch environments/{dev.tfvars,stage.tfvars,prod.tfvars}

# # Create scripts directory and shell scripts
# mkdir -p scripts
# touch scripts/{deploy.sh,destroy.sh}

# # Create root level files
# touch {README.md,.gitignore,.gitlab-ci.yml}

# # Make shell scripts executable
# chmod +x scripts/{deploy.sh,destroy.sh}

# echo "Project structure for lexy-chatbot has been created successfully!"
