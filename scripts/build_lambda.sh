#!/bin/bash
# scripts/build_lambda.sh

set -e

# Navigate to the lambda function directory
cd lambda/lex_main_handler

# Clean any existing artifacts
rm -f main main.zip

go mod tidy

# Build the Go binary
GOOS=linux GOARCH=amd64 go build -o main main.go


# Create the bootstrap file
echo '#!/bin/sh' > bootstrap
echo './main' >> bootstrap
chmod +x bootstrap

# Zip the binary
zip main.zip main bootstrap

# Clean up the binary
rm main bootstrap

cd ../../

echo "Lambda function built and zipped successfully."
