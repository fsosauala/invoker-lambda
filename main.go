package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	lambdaService "github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaInvocationTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

func HandleRequest(ctx context.Context, data events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	lambdaDestination := os.Getenv("destinationlambda")
	if lambdaDestination == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"no arn found"}`,
		}, nil
	}
	fmt.Println(lambdaDestination)
	region := os.Getenv("AWS_REGION")
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"cannot create config"}`,
		}, nil
	}
	lambdaClient := lambdaService.NewFromConfig(cfg)
	bodyToSend, err := json.Marshal(data)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error":"cannot marshal data"}`,
		}, nil
	}
	err = invokeLambda(ctx, lambdaDestination, lambdaClient, bodyToSend)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body:            fmt.Sprintf(`{"error":"%s"}`, err.Error()),
			IsBase64Encoded: false,
		}, nil
	}
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusAccepted,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"yeah":"la muneca fea"}`,
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}

func invokeLambda(ctx context.Context, lambdaDestination string, lambdaClient *lambdaService.Client, bodyToSend []byte) error {
	timeAsync := time.Now().UTC()
	output, err := lambdaClient.Invoke(
		ctx,
		&lambdaService.InvokeInput{
			FunctionName:   &lambdaDestination,
			Payload:        bodyToSend,
			InvocationType: lambdaInvocationTypes.InvocationTypeEvent,
		},
	)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error on async")
	}
	fmt.Printf("time since async call: %v\n", time.Since(timeAsync))
	fmt.Println("hehe async", string(output.Payload))

	timeSync := time.Now().UTC()
	output, err = lambdaClient.Invoke(
		ctx,
		&lambdaService.InvokeInput{
			FunctionName: &lambdaDestination,
			Payload:      bodyToSend,
		},
	)
	if err != nil {
		fmt.Println(err)
		return fmt.Errorf("error on sync")
	}
	fmt.Printf("time since sync call: %v\n", time.Since(timeSync))
	fmt.Println("hehe sync", string(output.Payload))

	return nil
}
