package main

import (
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

func run() {
	PostToSlack()
}

func main() {
	if os.Getenv("DEBUG") == "1" {
		run()
	} else {
		lambda.Start(run)
	}
}
