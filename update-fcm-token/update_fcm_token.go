package main

import (
	"../apimodel"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"strings"
	"github.com/ringoid/commons"
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
)

func init() {
	apimodel.InitLambdaVars("update-fcm-token-push")
}

func handler(ctx context.Context, request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
	lc, _ := lambdacontext.FromContext(ctx)

	userAgent := request.Headers["user-agent"]
	if strings.HasPrefix(userAgent, "ELB-HealthChecker") {
		return commons.NewServiceResponse("{}"), nil
	}

	if request.HTTPMethod != "POST" {
		return commons.NewWrongHttpMethodServiceResponse(), nil
	}
	sourceIp := request.Headers["x-forwarded-for"]

	apimodel.Anlogger.Debugf(lc, "update_fmc_token.go : handle request %v", request)

	appVersion, isItAndroid, ok, errStr := commons.ParseAppVersionFromHeaders(request.Headers, apimodel.Anlogger, lc)
	if !ok {
		apimodel.Anlogger.Errorf(lc, "update_fmc_token.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	reqParam, ok := parseParams(request.Body, lc)
	if !ok {
		errStr := commons.WrongRequestParamsClientError
		apimodel.Anlogger.Errorf(lc, "update_fmc_token.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	userId, ok, _, errStr := commons.CallVerifyAccessToken(appVersion, isItAndroid, reqParam.AccessToken, apimodel.InternalAuthFunctionName, apimodel.AwsLambdaClient, apimodel.Anlogger, lc)
	if !ok {
		apimodel.Anlogger.Errorf(lc, "update_fmc_token.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	ok, errStr = saveDeviceToken(reqParam.DeviceToken, userId, lc)
	if !ok {
		apimodel.Anlogger.Errorf(lc, "update_fmc_token.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	event := commons.NewDeviceTokenRegisteredEvent(userId, reqParam.DeviceToken, sourceIp, isItAndroid)
	commons.SendAnalyticEvent(event, userId, apimodel.DeliveryStreamName, apimodel.AwsDeliveryStreamClient, apimodel.Anlogger, lc)

	apimodel.Anlogger.Infof(lc, "update_fmc_token.go : successfully update device token [%s] for userId [%s]",
		reqParam.DeviceToken, userId)
	return commons.NewServiceResponse(string("{}")), nil
}

//return ok and error string
func saveDeviceToken(deviceToken, userId string, lc *lambdacontext.LambdaContext) (bool, string) {
	apimodel.Anlogger.Debugf(lc, "update_fmc_token.go : save device token [%s] for userId [%s]",
		deviceToken, userId)

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#deviceDeviceToken": aws.String(commons.DeviceTokenColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":deviceDeviceTokenV": {
				S: aws.String(deviceToken),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			commons.UserIdColumnName: {
				S: aws.String(userId),
			},
		},
		TableName:        aws.String(apimodel.TokenTableName),
		UpdateExpression: aws.String("SET #deviceDeviceToken = :deviceDeviceTokenV"),
	}

	_, err := apimodel.AwsDynamoDbClient.UpdateItem(input)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "update_fmc_token.go : error update device token [%s] for userId [%s] : %v",
			deviceToken, userId, err)
		return false, commons.InternalServerError
	}

	apimodel.Anlogger.Debugf(lc, "update_fmc_token.go : successfully update device token [%s] for userId [%s]",
		deviceToken, userId)
	return true, ""
}

func parseParams(params string, lc *lambdacontext.LambdaContext) (*apimodel.UpdateTokenRequest, bool) {
	var req apimodel.UpdateTokenRequest
	err := json.Unmarshal([]byte(params), &req)

	if err != nil {
		apimodel.Anlogger.Errorf(lc, "update_fmc_token.go : error unmarshal required params from the string %s : %v", params, err)
		return nil, false
	}

	if req.AccessToken == "" || req.DeviceToken == "" {
		apimodel.Anlogger.Errorf(lc, "update_fmc_token.go : one of the required params are empty, request [%v]", req)
		return nil, false
	}

	return &req, true
}

func main() {
	basicLambda.Start(handler)
}
