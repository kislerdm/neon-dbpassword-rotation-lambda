module github.com/kislerdm/aws-lambda-secret-rotation/plugin/neon

go 1.19

require (
	github.com/aws/aws-lambda-go v1.37.0
	github.com/aws/aws-sdk-go-v2/config v1.18.17
	github.com/aws/aws-sdk-go-v2/service/secretsmanager v1.18.1
	github.com/kislerdm/aws-lambda-secret-rotation v0.1.1
	github.com/kislerdm/neon-sdk-go v0.1.4
	github.com/lib/pq v1.10.7
)

require (
	github.com/aws/aws-sdk-go-v2 v1.17.6 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.13.17 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.24 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.31 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.9.24 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.12.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.14.5 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.18.6 // indirect
	github.com/aws/smithy-go v1.13.5 // indirect
)

replace github.com/kislerdm/aws-lambda-secret-rotation => ../..
