package main

import "github.com/aws/aws-lambda-go/lambda"

func main() {
	lambda.Start(handler)
}

func handler() error {
	// TODO: Management slash command
	return nil
}
