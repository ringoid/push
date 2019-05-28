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
		var aTask commons.PushObject
		err := json.Unmarshal([]byte(body), &aTask)
		if err != nil {
			apimodel.Anlogger.Errorf(lc, "internal_handle_task.go : error unmarshal body [%s] to commons.PushObject : %v", body, err)
			return errors.New(fmt.Sprintf("error unmarshal body %s : %v", body, err))
		}
		apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : handle record %v", aTask)

		onlineTime := aTask.LastOnlineTime
		period := int64(-1)
		if aTask.PushType == commons.OnceDayPushType {
			onlineTime = int64(-1)
			period = apimodel.MaxPeriodDefault
		}

		canWeSent, wasRequestOk, needToRetry := false, false, false
		var errStr string
		for {
			canWeSent, wasRequestOk, needToRetry, errStr = apimodel.CanPushTypeBeSent(aTask.UserId, aTask.PushType, onlineTime, period, lc)
			if wasRequestOk {
				break
			}
			if !wasRequestOk && needToRetry {
				continue
			}
			return errors.New(errStr)
		}
		var pushWasSentEvent *commons.PushWasSentToUser
		if canWeSent {
			//send notification push with optional data part
			pushWasSentEvent = commons.NewPushWasSentToUser(aTask.UserId, aTask.PushType)
			if aTask.PushType == commons.OnceDayPushType {
				ok, errStr := commons.SendCommonEvent(pushWasSentEvent, aTask.UserId, apimodel.CommonStreamName, aTask.UserId, apimodel.AwsKinesisStreamClient, apimodel.Anlogger, lc)
				if !ok {
					return errors.New(errStr)
				}
			}

			err = sendSpecialPush(aTask, false, lc)
			if err != nil {
				return err
			}
			pushCounter++
		} else {
			//send data push in this case
			pushWasSentEvent = commons.NewDataPushWasSentToUser(aTask.UserId, aTask.PushType)
			err = sendSpecialPush(aTask, true, lc)
			if err != nil {
				return err
			}
			dataPushCounter++
		}

		commons.SendAnalyticEvent(pushWasSentEvent, aTask.UserId, apimodel.DeliveryStreamName, apimodel.AwsDeliveryStreamClient, apimodel.Anlogger, lc)
	}

	apimodel.Anlogger.Debugf(lc, "internal_handle_task.go : successfully complete handle push requests with [%d] records and send [%d] pushes and [%d] data pushes",
		len(event.Records), pushCounter, dataPushCounter)
	return nil
}

func main() {
	basicLambda.Start(handler)
}
