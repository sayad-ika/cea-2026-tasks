# craftsbite-backend

Discord slash-command bot for managing daily meal participation, work locations, and team headcount — built on AWS Lambda + DynamoDB with native Go handlers.

## Architecture

Two Lambda tiers communicate over the AWS Lambda invoke API:

| Component           | Path             | Role                                                                                   |
| ------------------- | ---------------- | -------------------------------------------------------------------------------------- |
| Router Lambda       | `cmd/router/`    | Verifies Ed25519 signature, resolves caller identity, dispatches to per-command Lambda |
| Per-command Lambdas | `cmd/<command>/` | Handle a single slash command, reply via Discord followup REST call                    |

## Quick start (local)

```bash
# 1. Copy env template
cp .env.example .env

# 2. Start local DynamoDB
docker-compose up -d

# 3. Create tables + seed data
go run ./scripts/create_tables.go
go run ./scripts/seed.go
```

## Dependencies

| Package                                                         | Version  | Purpose                  |
| --------------------------------------------------------------- | -------- | ------------------------ |
| `github.com/aws/aws-lambda-go`                                  | v1.53.0  | Lambda handler + events  |
| `github.com/aws/aws-sdk-go-v2`                                  | v1.41.3  | AWS SDK core             |
| `github.com/aws/aws-sdk-go-v2/config`                           | v1.32.11 | AWS config loading       |
| `github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue`  | v1.20.34 | DynamoDB attribute codec |
| `github.com/aws/aws-sdk-go-v2/service/dynamodb`                 | v1.56.1  | DynamoDB client          |
| `github.com/aws/aws-sdk-go-v2/service/lambda`                   | v1.88.2  | Lambda invoke client     |
| `github.com/aws/aws-sdk-go-v2/service/ssm`                      | v1.68.2  | SSM parameter store      |
| `google.golang.org/api`                                         | v0.271.0 | Google Chat API client   |

Go version: **1.25.0**
