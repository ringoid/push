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
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"regexp"
)

func init() {
	apimodel.InitLambdaVars("update-token-push")
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

	apimodel.Anlogger.Debugf(lc, "update.go : handle request %v", request)

	appVersion, isItAndroid, ok, errStr := commons.ParseAppVersionFromHeaders(request.Headers, apimodel.Anlogger, lc)
	if !ok {
		apimodel.Anlogger.Errorf(lc, "update.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	reqParam, ok := parseParams(request.Body, lc)
	if !ok {
		errStr := commons.WrongRequestParamsClientError
		apimodel.Anlogger.Errorf(lc, "update.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	userId, ok, _, errStr := commons.CallVerifyAccessToken(appVersion, isItAndroid, reqParam.AccessToken, apimodel.InternalAuthFunctionName, apimodel.AwsLambdaClient, apimodel.Anlogger, lc)
	if !ok {
		apimodel.Anlogger.Errorf(lc, "update.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	if isItAndroid && apimodel.PlatformApplicationArnAndroid == "na" {
		apimodel.Anlogger.Infof(lc, "update.go : there is no PlatformApplicationArnAndroid, skip the call")
		return commons.NewServiceResponse(string("{}")), nil
	} else if !isItAndroid && apimodel.PlatformApplicationArnIos == "na" {
		apimodel.Anlogger.Infof(lc, "update.go : there is no PlatformApplicationArnIos, skip the call")
		return commons.NewServiceResponse(string("{}")), nil
	}

	platformArn, ok, errStr := getPlatformEndPointArnByDeviceToken(reqParam.DeviceToken, userId, isItAndroid, lc)
	if !ok {
		apimodel.Anlogger.Errorf(lc, "update.go : return %s to client", errStr)
		return commons.NewServiceResponse(errStr), nil
	}

	updatedNeeded := false
	createdNeeded := false
	if len(platformArn) == 0 {
		createdNeeded = true
	}

	if createdNeeded {
		platformArn, ok, errStr = createPlatformEndpoint(reqParam.DeviceToken, userId, isItAndroid, lc)
		if !ok {
			apimodel.Anlogger.Errorf(lc, "update.go : return %s to client", errStr)
			return commons.NewServiceResponse(errStr), nil
		}
	}

	input := &sns.GetEndpointAttributesInput{
		EndpointArn: aws.String(platformArn),
	}

	result, err := apimodel.AwsSnsClient.GetEndpointAttributes(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sns.ErrCodeNotFoundException:
				createdNeeded = true
			default:
				apimodel.Anlogger.Errorf(lc, "update.go : error get endpoint attributes for arn [%s] and userId [%s] : %v", platformArn, userId, aerr)
				apimodel.Anlogger.Errorf(lc, "update.go : return %s to client", commons.InternalServerError)
				return commons.NewServiceResponse(commons.InternalServerError), nil
			}
		} else {
			apimodel.Anlogger.Errorf(lc, "update.go : error get endpoint attributes for arn [%s] and userId [%s] : %v", platformArn, userId, err)
			apimodel.Anlogger.Errorf(lc, "update.go : return %s to client", commons.InternalServerError)
			return commons.NewServiceResponse(commons.InternalServerError), nil
		}
	} else {
		var actuallToken string
		if result.Attributes["Token"] != nil {
			actuallToken = *result.Attributes["Token"]
		}

		var actuallState string
		if result.Attributes["Enabled"] != nil {
			actuallState = *result.Attributes["Enabled"]
		}

		if actuallToken != reqParam.DeviceToken || actuallState != "true" {
			updatedNeeded = true
		}
	}

	if createdNeeded {
		platformArn, ok, errStr = createPlatformEndpoint(reqParam.DeviceToken, userId, isItAndroid, lc)
		if !ok {
			apimodel.Anlogger.Errorf(lc, "update.go : return %s to client", errStr)
			return commons.NewServiceResponse(errStr), nil
		}
	}

	if updatedNeeded {
		apimodel.Anlogger.Errorf(lc, "update.go : endpoint attributes need to be updated for platform endpoint [%s] and userId [%s]", platformArn, userId)
	}

	event := commons.NewDeviceTokenRegisteredEvent(userId, reqParam.DeviceToken, sourceIp, isItAndroid)
	commons.SendAnalyticEvent(event, userId, apimodel.DeliveryStreamName, apimodel.AwsDeliveryStreamClient, apimodel.Anlogger, lc)

	apimodel.Anlogger.Infof(lc, "update.go : successfully update platform endpoint [%s] for userId [%s] with device token [%s]",
		platformArn, userId, reqParam.DeviceToken)
	return commons.NewServiceResponse(string("{}")), nil
}

//return platform arn, ok and errStr
func createPlatformEndpoint(deviceToken, userId string, isItAndroid bool, lc *lambdacontext.LambdaContext) (string, bool, string) {
	apimodel.Anlogger.Debugf(lc, "update.go : create platform endpoint for device token [%s], is it android [%v] for userId [%s]",
		deviceToken, isItAndroid, userId)
	appArn := apimodel.PlatformApplicationArnIos
	if isItAndroid {
		appArn = apimodel.PlatformApplicationArnAndroid
	}
	input := &sns.CreatePlatformEndpointInput{
		PlatformApplicationArn: aws.String(appArn),
		Token:                  aws.String(deviceToken),
	}
	var endpointArn string
	result, err := apimodel.AwsSnsClient.CreatePlatformEndpoint(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case sns.ErrCodeInvalidParameterException:
				re := regexp.MustCompile(`(.*Endpoint) (\()(arn.+)(\)) (already exists with the same token)`)
				matches := re.FindStringSubmatch(aerr.Message())
				if len(matches) != 6 {
					apimodel.Anlogger.Errorf(lc, "update.go : error create platform endpoint for device token [%s], is it android [%v] for userId [%s] : %v",
						deviceToken, isItAndroid, userId, aerr)
					return "", false, commons.InternalServerError
				}
				endpointArn = matches[3]
			default:
				apimodel.Anlogger.Errorf(lc, "update.go : error create platform endpoint for device token [%s], is it android [%v] for userId [%s] : %v",
					deviceToken, isItAndroid, userId, aerr)
				return "", false, commons.InternalServerError
			}
		} else {
			apimodel.Anlogger.Errorf(lc, "update.go : error create platform endpoint for device token [%s], is it android [%v] for userId [%s] : %v",
				deviceToken, isItAndroid, userId, err)
			return "", false, commons.InternalServerError
		}
	} else {
		endpointArn = *result.EndpointArn
	}

	ok, errStr := saveEndpointArn(deviceToken, userId, endpointArn, isItAndroid, lc)
	if !ok {
		return "", false, errStr
	}

	apimodel.Anlogger.Debugf(lc, "update.go : successfully create platform endpoint for device token [%s], is it android [%v] for userId [%s]",
		deviceToken, isItAndroid, userId)

	return endpointArn, true, ""
}

//return ok and error string
func saveEndpointArn(deviceToken, userId, endpointArn string, isItAndroid bool, lc *lambdacontext.LambdaContext) (bool, string) {
	apimodel.Anlogger.Debugf(lc, "update.go : save endpoint arn [%s] by device token [%s], is it android [%v] for userId [%s]",
		endpointArn, deviceToken, isItAndroid, userId)
	os := commons.IOSOperationalSystemName
	if isItAndroid {
		os = commons.AndroidOperationalSystemName
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#platformArn": aws.String(commons.PlatformEndpointArnColumnName),
			"#userId":      aws.String(commons.UserIdColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":platformArnV": {
				S: aws.String(endpointArn),
			},
			":userIdV": {
				S: aws.String(userId),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			commons.DeviceTokenColumnName: {
				S: aws.String(deviceToken),
			},
			commons.OSColumnName: {
				S: aws.String(os),
			},
		},
		TableName:        aws.String(apimodel.TokenTableName),
		UpdateExpression: aws.String("SET #platformArn = :platformArnV, #userId = :userIdV"),
	}

	_, err := apimodel.AwsDynamoDbClient.UpdateItem(input)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "update.go : error update endpoint arn [%s] by device token [%s], is it android [%v] for userId [%s] : %v",
			endpointArn, deviceToken, isItAndroid, userId, err)
		return false, commons.InternalServerError
	}

	apimodel.Anlogger.Debugf(lc, "update.go : successfully update endpoint arn [%s] by device token [%s], is it android [%v] for userId [%s] : %v",
		endpointArn, deviceToken, isItAndroid, userId, err)
	return true, ""
}

//return platform arn, ok and error string
func getPlatformEndPointArnByDeviceToken(deviceToken, userId string, isItAndroid bool, lc *lambdacontext.LambdaContext) (string, bool, string) {
	apimodel.Anlogger.Debugf(lc, "update.go : get platform arn by device token [%s], is it android [%v] for userId [%s]", deviceToken, isItAndroid, userId)
	os := commons.IOSOperationalSystemName
	if isItAndroid {
		os = commons.AndroidOperationalSystemName
	}

	input := &dynamodb.GetItemInput{
		Key:
		map[string]*dynamodb.AttributeValue{
			commons.DeviceTokenColumnName: {
				S: aws.String(deviceToken),
			},
			commons.OSColumnName: {
				S: aws.String(os),
			},
		},
		ConsistentRead: aws.Bool(true),
		TableName:      aws.String(apimodel.TokenTableName),
	}

	result, err := apimodel.AwsDynamoDbClient.GetItem(input)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "update.go : error get platform arn by device token [%s], is it android [%v] for userId [%s] : %v", deviceToken, isItAndroid, userId, err)
		return "", false, commons.InternalServerError
	}

	if len(result.Item) == 0 {
		apimodel.Anlogger.Debugf(lc, "update.go : there is no platform arn for device token [%s], is it android [%v] for userId [%s]", deviceToken, isItAndroid, userId)
		return "", true, ""
	}

	platformArn := *result.Item[commons.PlatformEndpointArnColumnName].S
	currentUserId := *result.Item[commons.UserIdColumnName].S
	if currentUserId != userId {
		ok, errStr := saveEndpointArn(deviceToken, userId, platformArn, isItAndroid, lc)
		if !ok {
			return "", false, errStr
		}
	}
	apimodel.Anlogger.Debugf(lc, "update.go : successfully got platform arn [%s] for device token [%s], is it android [%v] for userId [%s]",
		platformArn, deviceToken, isItAndroid, userId)

	return platformArn, true, ""
}

func parseParams(params string, lc *lambdacontext.LambdaContext) (*apimodel.UpdateTokenRequest, bool) {
	var req apimodel.UpdateTokenRequest
	err := json.Unmarshal([]byte(params), &req)

	if err != nil {
		apimodel.Anlogger.Errorf(lc, "update.go : error unmarshal required params from the string %s : %v", params, err)
		return nil, false
	}

	if req.AccessToken == "" || req.DeviceToken == "" {
		apimodel.Anlogger.Errorf(lc, "update.go : one of the required params are empty, request [%v]", req)
		return nil, false
	}

	return &req, true
}

func main() {
	basicLambda.Start(handler)
}
