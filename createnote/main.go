package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
type Response events.APIGatewayProxyResponse

type note struct {
	UserID  string `json:"userId"`
	NoteID  string `json:"noteId"`
	Content string `json:"content"`
}

// CreateNote is our lambda handler invoked by the `lambda.Start` function call
func CreateNote(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	var buf bytes.Buffer

	id, err := uuid.NewUUID()
	if err != nil {
		return Response{StatusCode: 404}, err
	}

	cognitoIdentityID := request.RequestContext.Identity.CognitoIdentityID
	myNote := note{
		UserID:  cognitoIdentityID,
		NoteID:  id.String(),
		Content: request.Body,
	}

	dattr, err := dynamodbattribute.MarshalMap(myNote)
	if err != nil {
		return Response{StatusCode: 404}, err
	}

	svc := dynamodb.New(session.Must(session.NewSession()))
	input := &dynamodb.PutItemInput{
		Item:      dattr,
		TableName: aws.String(os.Getenv("tableName")),
	}
	_, err = svc.PutItem(input)
	if err != nil {
		return Response{StatusCode: 404}, err
	}

	body, err := json.Marshal(map[string]interface{}{
		"message": "Note created successfully!",
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}

	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      201,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type":           "application/json",
			"X-MyCompany-Func-Reply": "notes-handler",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(CreateNote)
}
