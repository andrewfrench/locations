package util

import (
	"fmt"

	"github.com/mmcloughlin/geohash"

	"github.com/andrewfrench/owntracks-go/pkg/owntracks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func ToAttributeValueMap(l *owntracks.Location) (map[string]*dynamodb.AttributeValue, error) {
	avm := map[string]*dynamodb.AttributeValue{}

	id, err := GenerateId(l)
	if err != nil {
		return nil, err
	}

	avm["id"] = &dynamodb.AttributeValue{S: &id}
	avm["type"] = &dynamodb.AttributeValue{S: &l.Type}
	avm["geoHash"] = &dynamodb.AttributeValue{S: aws.String(geohash.Encode(l.Latitude, l.Longitude))}
	avm["altitude"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.Altitude))}
	avm["latitude"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", l.Latitude))}
	avm["longitude"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", l.Longitude))}
	avm["reportedTimestamp"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.ReportedTimestamp))}
	avm["receivedTimestamp"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.ReceivedTimestamp))}
	avm["velocity"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.Velocity))}
	avm["course"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.Course))}
	avm["accuracy"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.Accuracy))}
	avm["battery"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.Battery))}
	avm["verticalAccuracy"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.VerticalAccuracy))}
	avm["pressure"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%f", l.Pressure))}
	avm["radius"] = &dynamodb.AttributeValue{N: aws.String(fmt.Sprintf("%d", l.Radius))}

	if l.TrackerID != "" {
		avm["trackerId"] = &dynamodb.AttributeValue{S: &l.TrackerID}
	}

	if l.ConnectivityStatus != "" {
		avm["connectivityStatus"] = &dynamodb.AttributeValue{S: &l.ConnectivityStatus}
	}

	if l.Topic != "" {
		avm["topic"] = &dynamodb.AttributeValue{S: &l.Topic}
	}

	if l.Trigger != "" {
		avm["trigger"] = &dynamodb.AttributeValue{S: &l.Trigger}
	}

	if len(l.InRegions) > 0 {
		s := make([]*string, 0)
		for _, r := range l.InRegions {
			if r != "" {
				s = append(s, &r)
			}
		}

		avm["inRegions"] = &dynamodb.AttributeValue{SS: s}
	}

	return avm, nil
}
