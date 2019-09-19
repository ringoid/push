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
	"github.com/aws/aws-sdk-go/aws/awserr"
	"firebase.google.com/go"
	"google.golang.org/api/option"
	"context"
	"firebase.google.com/go/messaging"
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
var ReadyForPushFunctionName string
var PushTaskQueue string
var AlreadySentPushTableName string
var CommonStreamName string

var FirebaseClient *messaging.Client

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

	Anlogger, err = commons.New(papertrailAddress, fmt.Sprintf("%s-%s", env, lambdaName), IsDebugLogEnabled)
	if err != nil {
		fmt.Errorf("lambda-initialization : service_common.go : error during startup : %v\n", err)
		os.Exit(1)
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : logger was successfully initialized")

	TokenTableName, ok = os.LookupEnv("FCM_TOKEN_TABLE_NAME")
	if !ok {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : env can not be empty FCM_TOKEN_TABLE_NAME")
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : start with FCM_TOKEN_TABLE_NAME = [%s]", TokenTableName)

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

	fireBaseCreds := commons.GetSecret(fmt.Sprintf("%s/Firebase/Credentials", env), "credentials", awsSession, Anlogger, nil)
	opt := option.WithCredentialsJSON([]byte(fireBaseCreds))
	fireBaseApp, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : error initialise firebase app : %v", err)
	}
	ctx := context.Background()
	FirebaseClient, err = fireBaseApp.Messaging(ctx)
	if err != nil {
		Anlogger.Fatalf(nil, "lambda-initialization : service_common.go : error initialise firebase client : %v", err)
	}
	Anlogger.Debugf(nil, "lambda-initialization : service_common.go : firebase client was successfully initialized")
}

//return can we send, was request ok, need to retry, and error string
func CanPushTypeBeSent(push commons.PushObject, oldestTimeForSendingPush, period int64, lc *lambdacontext.LambdaContext) (bool, bool, bool, string) {
	Anlogger.Debugf(lc, "service_common.go : check can we send push [%s] with oldestTimeForSendingPush [%v] and period [%v] for userId [%s]",
		push, oldestTimeForSendingPush, period, push.UserId)

	switch push.PushType {
	case commons.OnceDayPushType:
		Anlogger.Debugf(lc, "service_common.go : it's [%s] push for userId [%s]", push.PushType, push.UserId)
	case commons.NewLikePushType:
		if !push.NewLikeEnabled {
			Anlogger.Debugf(lc, "service_common.go : it's [%s] push, but new like push settings disabled, skip this push for userId [%s]",
				push.PushType, push.UserId)
			return false, true, false, ""
		}
	case commons.NewMatchPushType:
		if !push.NewMatchEnabled {
			Anlogger.Debugf(lc, "service_common.go : it's [%s] push, but new match push settings disabled, skip this push for userId [%s]",
				push.PushType, push.UserId)
			return false, true, false, ""
		}
	case commons.NewMessagePushType:
		if !push.NewMessageEnabled {
			Anlogger.Debugf(lc, "service_common.go : it's [%s] push, but new message push settings disabled, skip this push for userId [%s]",
				push.PushType, push.UserId)
			return false, true, false, ""
		}
	default:
		Anlogger.Errorf(lc, "service_common.go : Unsupported push type [%s] for userId [%s]", push.PushType, push.UserId)
		return false, false, false, commons.InternalServerError
	}

	if oldestTimeForSendingPush <= 0 && period <= 0 {
		Anlogger.Errorf(lc, "service_common.go : wrong params, when sending push for userId [%s], oldestTimeForSendingPush [%s], period [%s]",
			push.UserId, oldestTimeForSendingPush, period)
		return false, false, false, commons.InternalServerError
	}

	updateTime := commons.UnixTimeInMillis()
	oldestPossibleTimeForSendingPush := oldestTimeForSendingPush
	if oldestPossibleTimeForSendingPush <= 0 {
		oldestPossibleTimeForSendingPush = commons.UnixTimeInMillis() - period
	}

	pushId := push.UserId + "_" + push.PushType
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#updateTime": aws.String(commons.UpdatedTimeColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":oldestTimeV": {
				N: aws.String(fmt.Sprintf("%v", oldestPossibleTimeForSendingPush)),
			},
			":updateTimeV": {
				N: aws.String(fmt.Sprintf("%v", updateTime)),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			commons.UserIdColumnName: {
				S: aws.String(pushId),
			},
		},
		//if we use lastOnline time in params than this condition means that user
		//was online after last push with given type was sent
		ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s) OR %s < :oldestTimeV",
			commons.UpdatedTimeColumnName, commons.UpdatedTimeColumnName)),
		TableName:        aws.String(AlreadySentPushTableName),
		UpdateExpression: aws.String("SET #updateTime = :updateTimeV"),
	}

	_, err := AwsDynamoDbClient.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				Anlogger.Debugf(lc, "service_common.go : try to send push too often for userId [%s]", push.UserId)
				return false, true, false, ""
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				Anlogger.Warnf(lc, "service_common.go : warning, when sending push for userId [%s], need to retry : %v", push.UserId, aerr)
				return false, false, true, ""
			default:
				Anlogger.Errorf(lc, "service_common.go : error, try to send push for userId [%s] : %v", push.UserId, aerr)
				return false, false, false, commons.InternalServerError
			}
		}
		Anlogger.Errorf(lc, "service_common.go : error, try to send push for userId [%s] : %v", push.UserId, err)
		return false, false, false, commons.InternalServerError
	}

	Anlogger.Debugf(lc, "service_common.go : successfully check that we can send push for userId [%s] with updateTime [%v] and oldestPossibleTimeForSendingPush [%v]",
		push.UserId, updateTime, oldestPossibleTimeForSendingPush)
	return true, true, false, ""
}
