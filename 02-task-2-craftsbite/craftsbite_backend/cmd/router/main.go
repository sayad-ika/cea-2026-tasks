package main

import "github.com/aws/aws-lambda-go/lambda"

func main() {
	lambda.Start(handler)
}

func handler() error {
	// TODO: Ed25519 verification + identity resolution
	return nil
}
