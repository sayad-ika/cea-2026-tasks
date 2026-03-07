package main

import "github.com/aws/aws-lambda-go/lambda"

func main() {
	lambda.Start(handler)
}

func handler() error {
	// TODO: Self Cmd slash command
	return nil
}
