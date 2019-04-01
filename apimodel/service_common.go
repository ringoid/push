package apimodel

import (
	"github.com/ringoid/commons"
	"os"
	"github.com/aws/aws-sdk-go/aws/session"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/kinesis"
)

var Anlogger *commons.Logger
var AwsDynamoDbClient *dynamodb.DynamoDB
var AwsLambdaClient *lambda.Lambda
var AwsSnsClient *sns.SNS
var AwsSQSClient *sqs.SQS
var AwsDeliveryStreamClient *firehose.Firehose
var AwsKinesisStreamClient *kinesis.Kinesis

var TokenTableName string
var InternalAuthFunctionName string
var DeliveryStreamName string
var PlatformApplicationArnIos string
var PlatformApplicationArnAndroid string
var ReadyForPushFunctionName string
var PushTaskQueue string
var AlreadySentPushTableName string
var CommonStreamName string

func InitLambdaVars(lambdaName string) {
	var env string
	var ok bool
	var papertrailAddress string
	var err error
	var awsSession *session.Session

	env, ok = os.LookupEnv("ENV")
	if !ok {
		fmt.Printf("lambda-initialization : service_common.go : env can not be empty ENV\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : service_common.go : start with ENV = [%s]\n", env)

	papertrailAddress, ok = os.LookupEnv("PAPERTRAIL_LOG_ADDRESS")
	if !ok {
		fmt.Printf("lambda-initialization : service_common.go : env can not be empty PAPERTRAIL_LOG_ADDRESS\n")
		os.Exit(1)
	}
	fmt.Printf("lambda-initialization : service_common.go : start with PAPERTRAIL_LOG_ADDRESS = [%s]\n", papertrailAddress)

	Anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, lambdaName))
	if err != nil {
		fmt.Errorf("lambda-initialization : service_common.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : logger was successfully initialized")

	TokenTableName, ok = os.LookupEnv("TOKEN_TABLE_NAME")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty TOKEN_TABLE_NAME")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with TOKEN_TABLE_NAME = [%s]", TokenTableName)

	AlreadySentPushTableName, ok = os.LookupEnv("ALREADY_SENT_PUSH_TABLE_NAME")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty ALREADY_SENT_PUSH_TABLE_NAME")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with ALREADY_SENT_PUSH_TABLE_NAME = [%s]", AlreadySentPushTableName)

	InternalAuthFunctionName, ok = os.LookupEnv("INTERNAL_AUTH_FUNCTION_NAME")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty INTERNAL_AUTH_FUNCTION_NAME")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with INTERNAL_AUTH_FUNCTION_NAME = [%s]", InternalAuthFunctionName)

	ReadyForPushFunctionName, ok = os.LookupEnv("READY_FOR_PUSH_FUNCTION_NAME")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty READY_FOR_PUSH_FUNCTION_NAME")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with READY_FOR_PUSH_FUNCTION_NAME = [%s]", ReadyForPushFunctionName)

	DeliveryStreamName, ok = os.LookupEnv("DELIVERY_STREAM")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty DELIVERY_STREAM")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with DELIVERY_STREAM = [%s]", DeliveryStreamName)

	PlatformApplicationArnIos, ok = os.LookupEnv("PLATFORM_APPLICATION_ARN_IOS")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty PLATFORM_APPLICATION_ARN_IOS")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with PLATFORM_APPLICATION_ARN_IOS = [%s]", PlatformApplicationArnIos)

	PlatformApplicationArnAndroid, ok = os.LookupEnv("PLATFORM_APPLICATION_ARN_ANDROID")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty PLATFORM_APPLICATION_ARN_ANDROID")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with PLATFORM_APPLICATION_ARN_ANDROID = [%s]", PlatformApplicationArnAndroid)

	PushTaskQueue, ok = os.LookupEnv("PUSH_TASK_QUEUE")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty PUSH_TASK_QUEUE")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with PUSH_TASK_QUEUE = [%s]", PushTaskQueue)

	CommonStreamName, ok = os.LookupEnv("COMMON_STREAM")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty COMMON_STREAM")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with COMMON_STREAM = [%s]", CommonStreamName)

	awsSession, err = session.NewSession(aws.NewConfig().
		WithRegion(commons.Region).WithMaxRetries(commons.MaxRetries).
		WithLogger(aws.LoggerFunc(func(args ...interface{}) { Anlogger.AwsLog(args) })).WithLogLevel(aws.LogOff))
	if err != nil {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : error during initialization : %v", err)
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : aws session was successfully initialized")

	AwsDynamoDbClient = dynamodb.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : dynamodb client was successfully initialized")

	AwsLambdaClient = lambda.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : lambda client was successfully initialized")

	AwsSnsClient = sns.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : sns client was successfully initialized")

	AwsDeliveryStreamClient = firehose.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : firehose client was successfully initialized")

	AwsSQSClient = sqs.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : sqs client was successfully initialized")

	AwsKinesisStreamClient = kinesis.New(awsSession)
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : kinesis client was successfully initialized")
}

//return ok and error string
func PublishMessage(messageBody, title, userId string, lc *lambdacontext.LambdaContext) (bool, string) {
	Anlogger.Debugf(lc, "service_common.go : publish message [%s] to userId [%s]", messageBody, userId)
	endpointArn, os, ok, errStr := GetPlatformEndpointArnAndOs(userId, lc)
	if !ok {
		return false, errStr
	}
	if endpointArn == "" {
		Anlogger.Debugf(lc, "service_common.go : didn't send message to userId [%s], there is no endpoint arn", userId)
		return true, ""
	}
	msg := `{"GCM":"{\"notification\":{\"title\":\"` + title + `\",\"body\":\"` + messageBody + `\"}}"}`
	if os == commons.IOSOperationalSystemName {
		msg = `{"APNS":"{\"aps\":{\"alert\":\"` + messageBody + `\"}}"}`
	}
	input := &sns.PublishInput{
		Message: aws.String(msg),

		MessageStructure: aws.String("json"),
		TargetArn:        aws.String(endpointArn),
	}
	_, err := AwsSnsClient.Publish(input)
	if err != nil {
		Anlogger.Errorf(lc, "service_common.go : error publish message [%s] for userId [%s] : %v", msg, userId, err)
		return false, commons.InternalServerError
	}
	Anlogger.Debugf(lc, "service_common.go : successfully publish message [%s] to userId [%s]", msg, userId)
	return true, ""
}

//return platform arn, os, ok and error string
func GetPlatformEndpointArnAndOs(userId string, lc *lambdacontext.LambdaContext) (string, string, bool, string) {
	Anlogger.Debugf(lc, "service_common.go : get platform endpoint arn for userId [%s]", userId)
	input := &dynamodb.QueryInput{
		ExpressionAttributeNames: map[string]*string{
			"#userId": aws.String(commons.UserIdColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":userIdV": {
				S: aws.String(userId),
			},
		},

		KeyConditionExpression: aws.String("#userId = :userIdV"),
		TableName:              aws.String(TokenTableName),
		IndexName:              aws.String("userIdGSI"),
	}
	result, err := AwsDynamoDbClient.Query(input)
	if err != nil {
		Anlogger.Errorf(lc, "service_common.go : error query userIdGSI index for userId [%s] : %v", userId, err)
		return "", "", false, commons.InternalServerError
	}

	if len(result.Items) == 0 {
		Anlogger.Debugf(lc, "service_common.go : there is no platform endpoint for userId [%s]", userId)
		return "", "", true, ""
	}

	if len(result.Items) > 1 {
		Anlogger.Errorf(lc, "service_common.go : more than 1 endpoint arn for userId [%s], size [%d]", userId, len(result.Items))
		return "", "", false, commons.InternalServerError
	}

	endpointArn := *result.Items[0][commons.PlatformEndpointArnColumnName].S
	os := *result.Items[0][commons.OSColumnName].S
	Anlogger.Debugf(lc, "service_common.go : successfully get platform endpoint arn [%s] and os [%s] for userId [%s]", endpointArn, os, userId)

	return endpointArn, os, true, ""
}
