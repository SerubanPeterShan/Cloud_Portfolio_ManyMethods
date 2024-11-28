package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var dynamoClient *dynamodb.DynamoDB

type IPData struct {
	IP string `json:"ip"`
}

func saveIPHandler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Received request body: %s", event.Body)

	var data IPData
	err := json.Unmarshal([]byte(event.Body), &data)
	if err != nil {
		log.Printf("Error unmarshalling request body: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request body"}, nil
	}

	if data.IP == "" {
		log.Printf("IP address is empty")
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "IP address is empty"}, nil
	}

	av, err := dynamodbattribute.MarshalMap(data)
	if err != nil {
		log.Printf("DynamoDB marshal error: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Error marshalling data"}, nil
	}

	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	if tableName == "" {
		log.Printf("DynamoDB table name not set")
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "DynamoDB table name not set"}, nil
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	}

	_, err = dynamoClient.PutItemWithContext(ctx, input)
	if err != nil {
		log.Printf("DynamoDB insert error: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Error saving data to DynamoDB"}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 200, Body: "Data saved successfully"}, nil
}

func main() {
	// Initialize DynamoDB client
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ca-central-1"),
	}))
	dynamoClient = dynamodb.New(sess)

	// Start Lambda handler
	lambda.Start(saveIPHandler)
}
