package main

import (
	"context"
	"os/exec"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
)

type LambdaData struct {
	SubCommand string `json:"SubCommand"`
}

const binaryPath = "./sre-tooling"

// RunCommand is the AWS Lambda required handler. This function runs
// the sub-command provided in the LambdaData struct. See the AWS Lambda
// documentation[1].
//
// 1 - https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html
func RunCommand(ctx context.Context, data LambdaData) error {
	args := strings.Split(data.SubCommand, " ")
	cmd := exec.Command(binaryPath, args...)

	return cmd.Run()
}

func main() {
	lambda.Start(RunCommand)
}
