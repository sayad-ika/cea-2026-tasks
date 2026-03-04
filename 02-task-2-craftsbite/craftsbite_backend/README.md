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

| Package                                         | Version | Purpose                 |
| ----------------------------------------------- | ------- | ----------------------- |
| `github.com/aws/aws-lambda-go`                  | v1.47.0 | Lambda handler + events |
| `github.com/aws/aws-sdk-go-v2`                  | v1.32.0 | AWS SDK core            |
| `github.com/aws/aws-sdk-go-v2/config`           | v1.28.0 | AWS config loading      |
| `github.com/aws/aws-sdk-go-v2/service/dynamodb` | v1.36.0 | DynamoDB client         |
| `github.com/joho/godotenv`                      | v1.5.1  | `.env` file loading     |
| `go.uber.org/zap`                               | v1.27.0 | Structured logging      |
| `github.com/google/uuid`                        | v1.6.0  | UUID generation         |

Go version: **1.23**
