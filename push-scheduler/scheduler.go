package main

import (
	"../apimodel"
	"github.com/ringoid/commons"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"context"
	"fmt"
	"encoding/json"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"time"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/aws"
)

func init() {
	apimodel.InitLambdaVars("scheduler-push")
}

func handler(ctx context.Context, request interface{}) (string, error) {
	lc, _ := lambdacontext.FromContext(ctx)

	apimodel.Anlogger.Debugf(lc, "scheduler.go : start scheduling push events at [%v]", time.Now())
	start := commons.UnixTimeInMillis()

	pushCounter := 0
	skip := 0
	limit := 100
	for {
		resp, err := sendRequest(skip, limit, apimodel.ReadyForPushFunctionName, lc)
		if err != nil {
			return "", err
		}

		counter, err := sendPushToSqs(resp.Users, lc)
		if err != nil {
			return "", err
		}
		pushCounter += counter
		if resp.ResultCount < int64(limit) {
			apimodel.Anlogger.Debugf(lc, "scheduler.go : finish scheduling push events at [%v]", time.Now())
			apimodel.Anlogger.Infof(lc, "scheduler.go : successfully schedule [%d] pushes in [%v] millis", pushCounter, commons.UnixTimeInMillis()-start)
			//todo:send analytics event
			return "OK", nil
		}
		skip += limit
	}
}

func sendRequest(skip, limit int, functionName string, lc *lambdacontext.LambdaContext) (*commons.PushResponse, error) {
	apimodel.Anlogger.Debugf(lc, "scheduler.go : send request to [%s] to get push objects", functionName)
	request := commons.PushRequest{
		Skip:                int64(skip),
		Limit:               int64(limit),
		MaxPeriod:           apimodel.MaxPeriodDefault,
		OfflinePeriod:       apimodel.OfflinePeriodDefault,
		MinProfilesForMen:   apimodel.MinForMenDefault,
		MinProfilesForWomen: apimodel.MinForWomenDefault,
		MinH:                apimodel.MinH,
		MaxH:                apimodel.MaxH,
	}
	jsonBody, err := json.Marshal(request)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "scheduler.go : error marshaling request %v into json : %v", request, err)
		return nil, fmt.Errorf("error marshaling request %v into json : %v", request, err)
	}

	resp, err := apimodel.AwsLambdaClient.Invoke(&lambda.InvokeInput{FunctionName: aws.String(functionName), Payload: jsonBody})
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "scheduler.go : error invoke function [%s] to get push objects with body %s : %v",
			functionName, jsonBody, err)
		return nil, fmt.Errorf("scheduler.go : error invoke function [%s] to get push objects with body %s : %v",
			functionName, jsonBody, err)
	}

	if *resp.StatusCode != 200 {
		apimodel.Anlogger.Errorf(lc, "scheduler.go : status code = %d, response body %s for request %s (function name [%s])",
			*resp.StatusCode, string(resp.Payload), jsonBody, functionName)
		return nil, fmt.Errorf("scheduler.go : error invoke function [%s] with body %s : %v",
			functionName, jsonBody, err)
	}

	var response commons.PushResponse
	err = json.Unmarshal(resp.Payload, &response)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "scheduler.go : error unmarshaling response %s into commons.PushResponse : %v",
			string(resp.Payload), err)
		return nil, fmt.Errorf("error unmarshaling response %v into json : %v", string(resp.Payload), err)
	}

	apimodel.Anlogger.Debugf(lc, "scheduler.go : successfully receive [%v] push objects from [%s]", response, functionName)
	return &response, nil
}

func sendPushToSqs(source []commons.PushObject, lc *lambdacontext.LambdaContext) (int, error) {
	apimodel.Anlogger.Debugf(lc, "scheduler.go : put [%d] push objects into sqs", len(source))
	actualCounter := 0
	for _, each := range source {
		ok, _ := commons.SendAsyncTask(each, apimodel.PushTaskQueue, "admin", 0, apimodel.AwsSQSClient, apimodel.Anlogger, lc)
		if ok {
			actualCounter++
		}
	}
	apimodel.Anlogger.Debugf(lc, "scheduler.go : successfully put [%d] push objects into sqs", actualCounter)
	return actualCounter, nil
}

func main() {
	basicLambda.Start(handler)
}
