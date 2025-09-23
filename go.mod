module github.com/bhochhi/lexy

go 1.21.0

// For MVP, we avoid external deps beyond the stdlib. If you later integrate AWS Lex,
// add the AWS SDK v2 modules here (e.g., github.com/aws/aws-sdk-go-v2/service/lexruntimev2).

require (
	github.com/aws/aws-lambda-go v1.48.0 // indirect
	github.com/google/uuid v1.5.0 // indirect
)
