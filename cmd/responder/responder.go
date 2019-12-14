package main

import (
	"time"

	"github.com/andrewfrench/locations/pkg/env"
	"github.com/andrewfrench/locations/pkg/geojson"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mmcloughlin/geohash"
)

func main() {
	lambda.Start(func() (*geojson.GeoJson, error) {
		tableName, err := env.LocationsTable()
		if err != nil {
			return nil, err
		}

		sess := session.Must(session.NewSession())
		dCli := dynamodb.New(sess)

		geoJson := geojson.GeoJson{
			Type:     "FeatureCollection",
			Features: make([]geojson.Feature, 0),
		}

		var exclusiveStartKey map[string]*dynamodb.AttributeValue
		for {
			resp, err := dCli.Scan(&dynamodb.ScanInput{
				TableName:         aws.String(tableName),
				ExclusiveStartKey: exclusiveStartKey,
				AttributesToGet: []*string{
					aws.String("geoHash"),
				},
			})
			if err != nil {
				return nil, err
			}

			for _, item := range resp.Items {
				geoHash := *item["geoHash"].S
				lat, lng := geohash.DecodeCenter(geoHash)
				feature := geojson.Feature{
					Type: "Feature",
					Geometry: geojson.Geometry{
						Type:        "Point",
						Coordinates: []float32{float32(lng), float32(lat)},
					},
				}

				geoJson.Features = append(geoJson.Features, feature)
			}

			if resp.LastEvaluatedKey == nil {
				break
			} else {
				time.Sleep(time.Second)
			}

			exclusiveStartKey = resp.LastEvaluatedKey
		}

		return &geoJson, nil
	})
}
