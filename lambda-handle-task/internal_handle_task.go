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
)

func init() {
	apimodel.InitLambdaVars("internal-handle-task-push")
}

func handler(ctx context.Context, event events.SQSEvent) (error) {
	lc, _ := lambdacontext.FromContext(ctx)
	apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : start handle request with [%d] records", len(event.Records))
	var pushCounter int
	var dataPushCounter int
	for _, record := range event.Records {
		body := record.Body
		var pushTask commons.PushObject
		err := json.Unmarshal([]byte(body), &pushTask)
		if err != nil {
			apimodel.Anlogger.Errorf(lc, "internal_handle_task.go : error unmarshal body [%s] to commons.PushObject : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : handle record %v", pushTask)

		onlineTime := pushTask.LastOnlineTime
		period := int64(-1)
		if pushTask.PushType == commons.OnceDayPushType {
			onlineTime = int64(-1)
			period = apimodel.MaxPeriodDefault
		}

		canWeSent, wasRequestOk, needToRetry := false, false, false
		var errStr string
		for {
			canWeSent, wasRequestOk, needToRetry, errStr = apimodel.CanPushTypeBeSent(pushTask, onlineTime, period, lc)
			if wasRequestOk {
				break
			}
			if !wasRequestOk && needToRetry {
				continue
			}
			return errors.New(errStr)
		}
		var pushWasSentEvent *commons.PushWasSentToUser
		var sent bool
		if canWeSent {
			//send notification push with optional data part
			pushWasSentEvent = commons.NewPushWasSentToUser(pushTask.UserId, pushTask.PushType)
			if pushTask.PushType == commons.OnceDayPushType {
				ok, errStr := commons.SendCommonEvent(pushWasSentEvent, pushTask.UserId, apimodel.CommonStreamName, pushTask.UserId, apimodel.AwsKinesisStreamClient, apimodel.Anlogger, lc)
				if !ok {
					return errors.New(errStr)
				}
			}

			sent, err = sendSpecialPush(pushTask, false, lc)
			if err != nil && needThrowError(err) {
				return err
			}

			if sent {
				pushCounter++
			}

		} else {
			//send data push in this case
			pushWasSentEvent = commons.NewDataPushWasSentToUser(pushTask.UserId, pushTask.PushType)
			sent, err = sendSpecialPush(pushTask, true, lc)
			if err != nil && needThrowError(err) {
				return err
			}

			if sent {
				dataPushCounter++
			}
		}

		if sent {
			commons.SendAnalyticEvent(pushWasSentEvent, pushTask.UserId, apimodel.DeliveryStreamName, apimodel.AwsDeliveryStreamClient, apimodel.Anlogger, lc)
		}
	}

	apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : successfully complete handle push requests with [%d] records and send [%d] pushes and [%d] data pushes",
		len(event.Records), pushCounter, dataPushCounter)
	return nil
}

func needThrowError(err error) (bool) {
	strErr := fmt.Sprintf("%v", err)
	if strings.Contains(strErr, "registration-token-not-registered") ||
		strings.Contains(strErr, "invalid-argument") {
		return false
	}
	return true
}

func main() {
	basicLambda.Start(handler)
}
