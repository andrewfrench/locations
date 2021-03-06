package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/mmcloughlin/geohash"
)

const (
	Table             = "locations"
	Bucket            = "afrench-locations"
	PartialsKey       = "data/map.json"
	DigestKey         = "data/digest.json"
	ReportedTimestamp = "reportedTimestamp"
	GeoHash           = "geoHash"
	Region            = "us-west-2"
	ContentType       = "application/json"
	Limit             = 128
	AccuracyLimit     = 100
	Index             = "trackerId-reportedTimestamp-index"
)

type Digest struct {
	Size           int      `json:"size"`
	FirstTimestamp int      `json:"firstTimestamp"`
	LastTimestamp  int      `json:"lastTimestamp"`
	GeneratedAt    int      `json:"generatedAt"`
	Points         []*Point `json:"points"`
}

type Point struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}

func LoadCacheFromS3(dwn *s3manager.Downloader, dest *map[string]string) error {
	log.Infof("Loading cached data from S3")
	log.Infof("Downloading %s/%s", Bucket, PartialsKey)
	uniquesBuf := aws.NewWriteAtBuffer([]byte{})
	_, err := dwn.Download(uniquesBuf, &s3.GetObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(PartialsKey),
	})
	if err != nil {
		return err
	}
	log.Infof("Downloaded %s/%s", Bucket, PartialsKey)

	log.Infof("Parsing JSON")
	err = json.Unmarshal(uniquesBuf.Bytes(), dest)
	if err != nil {
		panic(err)
	}
	log.Infof("JSON loaded into points map (%d items)", len(*dest))

	return nil
}

func LoadRecentsFromDynamoDB(dyn *dynamodb.DynamoDB, fromTime string, dest *map[string]string) error {
	log.Infof("Loading data from DynamoDB %s/%s", Table, Index)
	log.Infof("Query size: %d", Limit)
	log.Infof("Query from: %s", fromTime)

	var next map[string]*dynamodb.AttributeValue = nil
	for {
		output, err := dyn.Query(&dynamodb.QueryInput{
			TableName:         aws.String(Table),
			ExclusiveStartKey: next,
			AttributesToGet: []*string{
				aws.String("geoHash"),
				aws.String("reportedTimestamp"),
				aws.String("accuracy"),
			},
			Limit:     aws.Int64(Limit),
			IndexName: aws.String(Index),
			KeyConditions: map[string]*dynamodb.Condition{
				"trackerId": {
					ComparisonOperator: aws.String("EQ"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							S: aws.String("8E"),
						},
					},
				},
				"reportedTimestamp": {
					ComparisonOperator: aws.String("GT"),
					AttributeValueList: []*dynamodb.AttributeValue{
						{
							N: &fromTime,
						},
					},
				},
			},
		})
		if err != nil {
			return err
		}

		pre := len(*dest)
		accuracyRejected := 0
		for _, item := range output.Items {
			accuracyValue, err := strconv.Atoi(*item["accuracy"].N)
			if err != nil {
				return err
			}

			if accuracyValue > AccuracyLimit {
				accuracyRejected++
				continue
			}

			(*dest)[*item[ReportedTimestamp].N] = *item[GeoHash].S
		}
		log.Infof("Added %d unique points, rejected %d due to accuracy (%d total)", len(*dest)-pre, accuracyRejected, len(*dest))

		next = output.LastEvaluatedKey
		if next == nil {
			break
		}

		time.Sleep(time.Second)
	}
	log.Infof("%d total unique points", len(*dest))

	return nil
}

func UploadCacheToS3(upl *s3manager.Uploader, data *map[string]string) error {
	log.Infof("Uploading cache to S3")

	log.Info("Marshalling data")
	outb, err := json.Marshal(*data)
	if err != nil {
		return err
	}

	log.Infof("Uploading %s/%s", Bucket, PartialsKey)
	_, err = upl.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(Bucket),
		Key:         aws.String(PartialsKey),
		Body:        bytes.NewBuffer(outb),
		ContentType: aws.String(ContentType),
	})
	if err != nil {
		return err
	}
	log.Infof("Upload complete")

	return nil
}

func BuildDigest(points map[string]string, dest *Digest) error {
	log.Infof("Building digest")
	dest.Size = len(points)
	dest.GeneratedAt = int(time.Now().Unix())
	for k := range points {
		t, err := strconv.Atoi(k)
		if err != nil {
			return err
		}

		dest.FirstTimestamp = t
		dest.LastTimestamp = t

		break
	}

	dest.Points = []*Point{}
	for k, v := range points {
		t, err := strconv.Atoi(k)
		if err != nil {
			return err
		}

		if t < dest.FirstTimestamp {
			dest.FirstTimestamp = t
		}

		if t > dest.LastTimestamp {
			dest.LastTimestamp = t
		}

		lat, lng := geohash.DecodeCenter(v)
		dest.Points = append(dest.Points, &Point{
			Lat: fmt.Sprintf("%0.4f", lat),
			Lng: fmt.Sprintf("%0.4f", lng),
		})
	}
	log.Info("Built digest")

	return nil
}

func UploadDigestToS3(upl *s3manager.Uploader, digest *Digest) error {
	log.Infof("Uploading digest to S3")

	log.Infof("Marshalling data")
	outb, err := json.Marshal(digest)
	if err != nil {
		return err
	}

	log.Infof("Uploading %s/%s", Bucket, DigestKey)
	_, err = upl.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(Bucket),
		Key:         aws.String(DigestKey),
		Body:        bytes.NewBuffer(outb),
		ContentType: aws.String(ContentType),
	})
	if err != nil {
		return err
	}
	log.Infof("Upload complete")

	return nil
}
