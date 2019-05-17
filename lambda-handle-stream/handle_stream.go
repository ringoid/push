package main

import (
	"context"
	basicLambda "github.com/aws/aws-lambda-go/lambda"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/ringoid/commons"
	"../apimodel"
	"fmt"
)

func init() {
	apimodel.InitLambdaVars("internal-handle-stream-push")
}

func handler(ctx context.Context, event events.KinesisEvent) (error) {
	lc, _ := lambdacontext.FromContext(ctx)

	apimodel.Anlogger.Debugf(lc, "handle_stream.go : start handle request with [%d] records", len(event.Records))

	for _, record := range event.Records {
		body := record.Kinesis.Data

		var aEvent commons.BaseInternalEvent
		err := json.Unmarshal(body, &aEvent)
		if err != nil {
			apimodel.Anlogger.Errorf(lc, "handle_stream.go : error unmarshal body [%s] to BaseInternalEvent : %v", string(body), err)
			if err != nil {
				apimodel.Anlogger.Errorf(lc, "handle_stream.go : skip record [%s]", string(body))
				continue
			}
		}
		apimodel.Anlogger.Debugf(lc, "handle_stream.go : handle record %v", aEvent)

		switch aEvent.EventType {
		case commons.NewUserLikeInternalEvent:
			err = sendLikeNotification(body, lc)
			if err != nil {
				apimodel.Anlogger.Errorf(lc, "handle_stream.go : skip record [%s]", string(body))
			}
		case commons.NewUserMatchInternalEvent:
			err = sendMatchNotification(body, lc)
			if err != nil {
				apimodel.Anlogger.Errorf(lc, "handle_stream.go : skip record [%s]", string(body))
			}
		case commons.NewUserMessageInternalEvent:
			err = sendMessageNotification(body, lc)
			if err != nil {
				apimodel.Anlogger.Errorf(lc, "handle_stream.go : skip record [%s]", string(body))
			}
		case commons.UserDeleteHimselfEvent:
			err = deleteUser(body, lc)
			if err != nil {
				apimodel.Anlogger.Errorf(lc, "handle_stream.go : skip record [%s]", string(body))
			}
		}

	}

	apimodel.Anlogger.Debugf(lc, "handle_stream.go : successfully complete handle request with [%d] records", len(event.Records))
	return nil
}

func unmarshal(body []byte, lc *lambdacontext.LambdaContext) (*commons.NewUserNotificationInternalEvent, error) {
	var event commons.NewUserNotificationInternalEvent
	err := json.Unmarshal(body, &event)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "handle_stream.go : error unmarshal body [%s] to BaseInternalEvent : %v", string(body), err)
		return nil, err
	}
	return &event, nil
}

func sendLikeNotification(body []byte, lc *lambdacontext.LambdaContext) (error) {
	event, err := unmarshal(body, lc)
	if err != nil {
		return err
	}
	push := commons.PushObject{
		UserId:         event.UserId,
		Sex:            event.Sex,
		Locale:         event.Locale,
		LastOnlineTime: event.LastOnlineTime,
		NewLikeCounter: 1,
		PushType:       commons.NewLikePushType,
	}
	err = sendPushToSqs(push, lc)
	return err
}

func sendMatchNotification(body []byte, lc *lambdacontext.LambdaContext) (error) {
	event, err := unmarshal(body, lc)
	if err != nil {
		return err
	}
	push := commons.PushObject{
		UserId:          event.UserId,
		Sex:             event.Sex,
		Locale:          event.Locale,
		LastOnlineTime:  event.LastOnlineTime,
		NewMatchCounter: 1,
		PushType:        commons.NewMatchPushType,
	}
	err = sendPushToSqs(push, lc)
	return err
}

func sendMessageNotification(body []byte, lc *lambdacontext.LambdaContext) (error) {
	event, err := unmarshal(body, lc)
	if err != nil {
		return err
	}
	push := commons.PushObject{
		UserId:            event.UserId,
		Sex:               event.Sex,
		Locale:            event.Locale,
		LastOnlineTime:    event.LastOnlineTime,
		NewMessageCounter: 1,
		PushType:          commons.NewMessagePushType,
	}
	err = sendPushToSqs(push, lc)
	return err
}

func sendPushToSqs(push commons.PushObject, lc *lambdacontext.LambdaContext) (error) {
	apimodel.Anlogger.Debugf(lc, "handle_stream.go : put [%v] push objects into sqs", push)
	ok, errStr := commons.SendAsyncTask(push, apimodel.PushTaskQueue, "admin", 0, apimodel.AwsSQSClient, apimodel.Anlogger, lc)
	if !ok {
		apimodel.Anlogger.Errorf(lc, "handle_stream.go : error sending push object to sqs : %s", errStr)
		return fmt.Errorf("%s", errStr)
	}
	apimodel.Anlogger.Debugf(lc, "handle_stream.go : successfully put [%v] push objects into sqs", push)
	return nil
}

func main() {
	basicLambda.Start(handler)
}
