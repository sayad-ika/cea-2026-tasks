# craftsbite-backend

Meal headcount planning system for office teams — Discord slash commands backed by Go + AWS Lambda + DynamoDB.

## Quick start

```bash
cp .env.example .env   # fill in required values
docker-compose up -d   # start local DynamoDB
make run               # start local server on :8080
```

## Build

```bash
make build   # produces bootstrap + function.zip (Linux amd64)
make test    # run all tests
```
