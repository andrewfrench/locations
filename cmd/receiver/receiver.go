package main

import (
	"context"
	"fmt"
	"time"

	"github.com/andrewfrench/locations/pkg/env"
	"github.com/andrewfrench/locations/pkg/util"

	"github.com/andrewfrench/owntracks-go/pkg/owntracks"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	fmt.Println("Entering Lambda")
	lambda.Start(func(ctx context.Context, le owntracks.Location) (string, error) {
		fmt.Printf("Message type: %s\n", le.Type)
		if le.Type != owntracks.TypeLocation {
			fmt.Printf("Message is not a location type\n")

			return "", nil
		}

		//accuracyLimit, err := env.AccuracyLimit()
		//if err != nil {
		//	return "", err
		//}

		//if le.Accuracy >= accuracyLimit {
		//	fmt.Printf("Reported accuracy (%d) meets or exceeds accuracy limit (%d), data will not be recorded\n", le.Accuracy, accuracyLimit)
		//
		//	return "", nil
		//}

		tableName, err := env.LocationsTable()
		if err != nil {
			return "", err
		}

		sess := session.Must(session.NewSession())
		dCli := dynamodb.New(sess)
		fmt.Println("Created DynamoDB client")

		le.ReceivedTimestamp = time.Now().Unix()
		avm, err := util.ToAttributeValueMap(&le)
		if err != nil {
			return "", err
		}
		fmt.Println("Created AttributeValueMap")

		_, err = dCli.PutItem(&dynamodb.PutItemInput{
			TableName: &tableName,
			Item:      avm,
		})
		for err != nil {
			_, err = dCli.PutItem(&dynamodb.PutItemInput{
				TableName: &tableName,
				Item:      avm,
			})

			time.Sleep(10 * time.Second)
		}

		fmt.Println("Put record")

		return "", nil
	})
	fmt.Println("Exiting Lambda")
}
