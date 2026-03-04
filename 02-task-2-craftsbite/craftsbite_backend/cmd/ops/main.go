package main

import "github.com/aws/aws-lambda-go/lambda"

func main() {
	lambda.Start(handler)
}

func handler() error {
	// TODO: Operation slash command
	return nil
}
