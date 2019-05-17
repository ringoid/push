package main

import (
	"../apimodel"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"fmt"
	"encoding/json"
	"errors"
	"github.com/ringoid/commons"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
)

func deleteUser(body []byte, lc *lambdacontext.LambdaContext) error {
	apimodel.Anlogger.Debugf(lc, "delete_user.go : handle event and delete user's token, body %s", string(body))
	var aEvent commons.UserCallDeleteHimselfEvent
	err := json.Unmarshal([]byte(body), &aEvent)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "delete_user.go : error unmarshal body [%s] to UserCallDeleteHimselfEvent: %v", string(body), err)
		return errors.New(fmt.Sprintf("error unmarshal body %s : %v", string(body), err))
	}

	input := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			commons.UserIdColumnName: {
				S: aws.String(aEvent.UserId),
			},
		},
		TableName: aws.String(apimodel.TokenTableName),
	}

	_, err = apimodel.AwsDynamoDbClient.DeleteItem(input)
	if err != nil {
		apimodel.Anlogger.Errorf(lc, "delete_user : error delete user with userId [%s] : %v", aEvent.UserId, err)
		return err
	}

	apimodel.Anlogger.Debugf(lc, "delete_user.go : successfully handle event and delete user's token, body %s", string(body))
	return nil
}
