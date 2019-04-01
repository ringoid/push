package main

import (
	"../apimodel"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
)

func init() {
	apimodel.InitLambdaVars("update-token-push")
}

type PublishRequest struct {
	UserId  string `json:"userId"`
	Message string `json:"message"`
}

func handler(ctx context.Context, request PublishRequest) (string, error) {
	lc, _ := lambdacontext.FromContext(ctx)
	ok, errStr := apimodel.PublishMessage(request.Message, "title", request.UserId, lc)
	if !ok {
		return errStr, nil
	}
	return "OK", nil
}

func main() {
	basicLambda.Start(handler)
}
