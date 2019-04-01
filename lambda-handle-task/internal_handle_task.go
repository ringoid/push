package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"../apimodel"
	"fmt"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"errors"
	"github.com/ringoid/commons"
	"strings"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
)

func init() {
	apimodel.InitLambdaVars("internal-handle-task-push")
}

func handler(ctx context.Context, event events.SQSEvent) (error) {
	lc, _ := lambdacontext.FromContext(ctx)
	apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : start handle request with [%d] records", len(event.Records))

	for _, record := range event.Records {
		body := record.Body
		var aTask commons.PushObject
		err := json.Unmarshal([]byte(body), &aTask)
		if err != nil {
			apimodel.Anlogger.Errorf(lc, "internal_handle_task.go : error unmarshal body [%s] to commons.PushObject : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : handle record %v", aTask)

		//todo:send push
		canWeSent, wasRequestOk, needToRetry := false, false, false
		var errStr string
		for {
			canWeSent, wasRequestOk, needToRetry, errStr = canPushBeSent(aTask.UserId, lc)
			if wasRequestOk {
				break
			}
			if !wasRequestOk && needToRetry {
				continue
			}
			return errors.New(errStr)
		}
		if canWeSent {

			pushWasSentEvent := commons.NewPushWasSentToUser(aTask.UserId, "base")

			commons.SendAnalyticEvent(pushWasSentEvent, aTask.UserId, apimodel.DeliveryStreamName, apimodel.AwsDeliveryStreamClient, apimodel.Anlogger, lc)

			ok, errStr := commons.SendCommonEvent(pushWasSentEvent, aTask.UserId, apimodel.CommonStreamName, aTask.UserId, apimodel.AwsKinesisStreamClient, apimodel.Anlogger, lc)
			if !ok {
				return errors.New(errStr)
			}

			ok, errStr = apimodel.PublishMessage(apimodel.MessageTexts[strings.ToLower(aTask.Locale)], "", aTask.UserId, lc)
			if !ok {
				apimodel.Anlogger.Errorf(lc, "internal_handle_task.go : error send push to userId [%s] : %s", aTask.UserId, errStr)
				return errors.New(fmt.Sprintf("error send push to userId [%s] : %v", aTask.UserId, errStr))
			}
		}
	}

	apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : successfully complete handle push requests with [%d] records", len(event.Records))
	return nil
}

//return can we send, was request ok, need to retry, and error string
func canPushBeSent(userId string, lc *lambdacontext.LambdaContext) (bool, bool, bool, string) {
	apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : check can we send push for userId [%s]", userId)
	currTime := commons.UnixTimeInMillis()
	alTime := commons.UnixTimeInMillis() - apimodel.MaxPeriodDefault

	apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : currTime [%v], alTime [%v]", currTime, alTime)

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeNames: map[string]*string{
			"#currTime": aws.String(commons.UpdatedTimeColumnName),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":alTimeV": {
				N: aws.String(fmt.Sprintf("%v", alTime)),
			},
			":currTimeV": {
				N: aws.String(fmt.Sprintf("%v", currTime)),
			},
		},
		Key: map[string]*dynamodb.AttributeValue{
			commons.UserIdColumnName: {
				S: aws.String(userId),
			},
		},
		ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(%s) OR %s < :alTimeV",
			commons.UpdatedTimeColumnName, commons.UpdatedTimeColumnName)),
		TableName:        aws.String(apimodel.AlreadySentPushTableName),
		UpdateExpression: aws.String("SET #currTime = :currTimeV"),
	}

	_, err := apimodel.AwsDynamoDbClient.UpdateItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : try to send push too often for userId [%s]", userId)
				return false, true, false, ""
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				apimodel.Anlogger.Warnf(lc, "internal_handle_task.go : warning, try to send push too often for userId [%s], need to retry : %v", userId, aerr)
				return false, false, true, ""
			default:
				apimodel.Anlogger.Errorf(lc, "internal_handle_task.go : error, try to send push for userId [%s] : %v", userId, aerr)
				return false, false, false, commons.InternalServerError
			}
		}
		apimodel.Anlogger.Errorf(lc, "internal_handle_task.go : error, try to send push for userId [%s] : %v", userId, err)
		return false, false, false, commons.InternalServerError
	}

	apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : successfully check that we can send push for userId [%s]", userId)
	return true, true, false, ""
}

func main() {
	basicLambda.Start(handler)
}
