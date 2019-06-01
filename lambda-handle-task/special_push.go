package main

import (
	"github.com/ringoid/commons"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"../apimodel"
	"strings"
	"fmt"
	"firebase.google.com/go/messaging"
	"context"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
	"errors"
)

//was push sent and error
func sendSpecialPush(push commons.PushObject, isItDataPush bool, lc *lambdacontext.LambdaContext) (bool, error) {
	token, ok, errStr := fetchToken(push.UserId, lc)
	if !ok {
		return false, errors.New(errStr)
	}

	if token == "" {
		apimodel.Anlogger.Debugf(lc, "special_push.go : there is no token for userId [%s], skipp push [%v]",
			push.UserId, push)
		return false, nil
	}

	var err error
	switch push.PushType {
	case commons.OnceDayPushType:
		err = sendOnceDayPush(push.UserId, token, push, lc)
	case commons.NewLikePushType:
		err = sendNewLikePush(push.UserId, token, push, isItDataPush, lc)
	case commons.NewMatchPushType:
		err = sendNewMatchPush(push.UserId, token, push, isItDataPush, lc)
	case commons.NewMessagePushType:
		err = sendNewMessagePush(push.UserId, token, push, isItDataPush, lc)
	default:
		err = fmt.Errorf("Unsupported push type [%s]", push.PushType)
	}

	if err != nil {
		apimodel.Anlogger.Errorf(lc, "special_push.go : error when send push for userId [%s], push [%v] : %v", push.UserId, push, err)
		return false, err
	}

	apimodel.Anlogger.Debugf(lc, "special_push.go : successfully send push for userId [%s], push [%v]",
		push.UserId, push)
	return true, nil
}

//return token, ok and error string
func fetchToken(userId string, lc *lambdacontext.LambdaContext) (string, bool, string) {
	apimodel.Anlogger.Debugf(lc, "special_push.go : get device token for userId [%s]", userId)

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			commons.UserIdColumnName: {
				S: aws.String(userId),
			},
		},
		TableName:      aws.String(apimodel.TokenTableName),
		ConsistentRead: aws.Bool(true),
	}

	result, err := apimodel.AwsDynamoDbClient.GetItem(input)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "special_push.go : error get device token for userId [%s] : %v", userId, err)
		return "", false, commons.InternalServerError
	}

	if len(result.Item) == 0 {
		apimodel.Anlogger.Debugf(lc, "special_push.go : there is no device token for userId [%s]", userId)
		return "", true, ""
	}

	deviceToken := *result.Item[commons.DeviceTokenColumnName].S
	apimodel.Anlogger.Debugf(lc, "special_push.go : successfully get device token [%s] for userId [%s]", deviceToken, userId)

	return deviceToken, true, ""
}

func createMessage(token, messageBody string, push commons.PushObject, isItDataPush bool) (*messaging.Message) {
	msg := &messaging.Message{
		Data: map[string]string{
			"type":           push.PushType,
			"oppositeUserId": push.OppositeUserId,
		},
		Token: token,
	}

	if !isItDataPush {
		msg.Notification = &messaging.Notification{
			Body: messageBody,
		}
		msg.Android = &messaging.AndroidConfig{
			CollapseKey: "new_like_message_collapse_key",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
				Tag:   "new_like_message",
			},
		}
		msg.APNS = &messaging.APNSConfig{
			Headers: map[string]string{"apns-collapse-id": "new_like_message_collapse_key"},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					//If you provide a Notification Content app extension, you can use this value to group your notifications together.
					ThreadID: "new_like_message_thread_id",
				},
			},
		}
	}
	return msg
}

func sendOnceDayPush(userId, token string, push commons.PushObject, lc *lambdacontext.LambdaContext) (error) {
	messageBody, ok := apimodel.NewPeopleMessageTexts[strings.ToLower(push.Locale)]
	if push.NewLikeCounter > 0 ||
		push.NewMatchCounter > 0 ||
		push.NewMessageCounter > 0 {

		messageBody, ok = apimodel.NewLmmDataMessageTexts[strings.ToLower(push.Locale)]
	}

	if !ok {
		messageBody = apimodel.NewPeopleMessageTexts["en"]
	}

	msg := &messaging.Message{
		Notification: &messaging.Notification{
			Body: messageBody,
		},
		Android: &messaging.AndroidConfig{
			CollapseKey: "one_day_message_collapse_key",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
				Tag:   "one_day_message",
			},
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{"apns-collapse-id": "one_day_message_collapse_key"},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Sound: "default",
					//If you provide a Notification Content app extension, you can use this value to group your notifications together.
					ThreadID: "one_day_message_thread_id",
				},
			},
		},
		Token: token,
	}

	return sendPush(userId, msg, lc)
}

func sendNewLikePush(userId, token string, push commons.PushObject, isItDataPush bool, lc *lambdacontext.LambdaContext) (error) {
	messageBody, ok := apimodel.NewLikeMessageTexts[strings.ToLower(push.Locale)]
	if !ok {
		messageBody = apimodel.NewLikeMessageTexts["en"]
	}

	msg := createMessage(token, messageBody, push, isItDataPush)

	return sendPush(userId, msg, lc)
}

func sendNewMatchPush(userId, token string, push commons.PushObject, isItDataPush bool, lc *lambdacontext.LambdaContext) (error) {
	messageBody, ok := apimodel.NewMatchMessageTexts[strings.ToLower(push.Locale)]
	if !ok {
		messageBody = apimodel.NewMatchMessageTexts["en"]
	}

	msg := createMessage(token, messageBody, push, isItDataPush)

	return sendPush(userId, msg, lc)
}

func sendNewMessagePush(userId, token string, push commons.PushObject, isItDataPush bool, lc *lambdacontext.LambdaContext) (error) {
	messageBody, ok := apimodel.NewMessageMessageTexts[strings.ToLower(push.Locale)]
	if !ok {
		messageBody = apimodel.NewMessageMessageTexts["en"]
	}

	msg := createMessage(token, messageBody, push, isItDataPush)

	return sendPush(userId, msg, lc)
}

func sendPush(userId string, msg *messaging.Message, lc *lambdacontext.LambdaContext) (error) {
	ctx := context.Background()
	_, err := apimodel.FirebaseClient.Send(ctx, msg)
	if err == nil {
		apimodel.Anlogger.Debugf(lc, "special_push.go : push [%v] for userId [%s] was successfully sent",
			msg, userId)
	}
	return err
}
