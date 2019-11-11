package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Response is of type APIGatewayProxyResponse since we're leveraging the
// AWS Lambda Proxy Request functionality (default behavior)
type Response events.APIGatewayProxyResponse

type Note struct {
	UserID  string `json:"userId"`
	NoteID  string `json:"noteId"`
	Content string `json:"content"`
}

//Event represents the input to the identification API
type Event struct {
	NoteId string `json:"note_id"`
}

// SearchNote is our lambda handler invoked by the `lambda.Start` function call
func SearchNote(ctx context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	apiEvent := Event{}

	if err := json.Unmarshal([]byte(request.Body), &apiEvent); err != nil {
		return Response{StatusCode: 404}, err
	}

	client := dynamodb.New(session.Must(session.NewSession()))

	// Dangerous use of GreaterThan without proper validation of input value (" " or * will return
	// everything from the table)
	queryFilter := expression.Name("noteId").GreaterThan(expression.Value(apiEvent.NoteId))

	queryExpression, err := expression.NewBuilder().WithFilter(queryFilter).Build()
	if err != nil {
		return Response{StatusCode: 404}, err
	}

	queryParameters := &dynamodb.ScanInput{
		TableName:                 aws.String(os.Getenv("tableName")),
		Select:                    aws.String("ALL_ATTRIBUTES"),
		ExpressionAttributeNames:  queryExpression.Names(),
		ExpressionAttributeValues: queryExpression.Values(),
		FilterExpression:          queryExpression.Filter(),
	}

	// Make the DynamoDB Query API call
	result, err := client.Scan(queryParameters)
	if err != nil {
		return Response{StatusCode: 404}, err
	}

	var notes []Note

	for _, i := range result.Items {
		note := Note{}
		err = dynamodbattribute.UnmarshalMap(i, &note)
		if err != nil {
			return Response{StatusCode: 404}, err
		}
		notes = append(notes, note)
	}

	body, err := json.Marshal(map[string]interface{}{
		"Notes": notes,
	})
	if err != nil {
		return Response{StatusCode: 404}, err
	}

	var buf bytes.Buffer

	json.HTMLEscape(&buf, body)

	resp := Response{
		StatusCode:      201,
		IsBase64Encoded: false,
		Body:            buf.String(),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	return resp, nil
}

func main() {
	lambda.Start(SearchNote)
}
