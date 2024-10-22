# Lexy Chatbot

This project implements a chatbot using Amazon Lex and AWS Lambda, managed with Terraform.

## Project Structure

The project is organized as follows:

```
lexy-chatbot/
├── README.md
├── .gitignore
├── .gitlab-ci.yml
├── terraform/
├── lambda/
├── environments/
└── scripts/
```

### Root Directory

- `README.md`: This file, containing project documentation.
- `.gitignore`: Specifies intentionally untracked files to ignore in Git.
- `.gitlab-ci.yml`: Configuration file for GitLab CI/CD pipelines.

### Terraform Directory

The `terraform/` directory contains all Terraform configurations for managing AWS resources.

```
terraform/
├── main.tf
├── variables.tf
├── outputs.tf
├── versions.tf
├── backend.tf
└── modules/
    ├── lambdas/
    ├── lex/
    └── iam/
```

- `main.tf`: The main Terraform configuration file.
- `variables.tf`: Defines input variables for the Terraform root module.
- `outputs.tf`: Specifies the outputs of the Terraform root module.
- `versions.tf`: Defines required versions for Terraform and providers.
- `backend.tf`: Configures the backend for storing Terraform state.
- `modules/`: Contains reusable Terraform modules:
    - `lambdas/`: Modules for Lambda functions.
    - `lex/`: Module for Amazon Lex configuration.
    - `iam/`: Module for IAM roles and policies.

### Lambda Directory

The `lambda/` directory contains the source code for AWS Lambda functions.

```
lambda/
└── lex_main_handler/
    └── main.go
```

- `lex_main_handler/`: Contains the main Lambda function that serves as the code hook for the Lex bot.

### Environments Directory

The `environments/` directory contains environment-specific Terraform variable files.

```
environments/
├── dev.tfvars
├── stage.tfvars
└── prod.tfvars
```

These files allow for environment-specific configurations:
- `dev.tfvars`: Variables for the development environment.
- `stage.tfvars`: Variables for the staging environment.
- `prod.tfvars`: Variables for the production environment.

### Scripts Directory

The `scripts/` directory contains utility scripts for managing the project.

```
scripts/
├── deploy.sh
├── destroy.sh
├── update_lambda.sh
└── build_lambda.sh
```

- `deploy.sh`: Script for deploying the infrastructure.
- `destroy.sh`: Script for tearing down the infrastructure.
- `update_lambda.sh`: Script for updating the Lambda functions.
- `build_lambda.sh`: Script for building the Lambda functions.


## Getting Started

- Navigate to the Terraform directory
`cd terraform`

- Create the S3 Bucket:

    ```bash
       aws s3api create-bucket --bucket lexy-chatbot-tf-state-bucket --region us-east-1
    ```

    Verify the Bucket:

    ```bash
       aws s3api list-buckets
    ```

- Run terraform init
```bash
terraform init
```



## Contributing


## License
