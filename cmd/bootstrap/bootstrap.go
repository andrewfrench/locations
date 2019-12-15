package main

import (
	"github.com/andrewfrench/locations/pkg/common"
	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func main() {
	cfg := &aws.Config{Region: aws.String(common.Region)}
	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	dyn := dynamodb.New(sess)
	upl := s3manager.NewUploader(sess)

	// Points are stored as a map of reportedTimestamp -> geoHash
	points := map[string]string{}

	// Bootstrapping requires pulling all points, start from time 0.
	fromTime := "0"
	err = common.LoadRecentsFromDynamoDB(dyn, fromTime, &points)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = common.UploadCacheToS3(upl, &points)
	if err != nil {
		log.Fatal(err.Error())
	}

	digest := &common.Digest{}
	err = common.BuildDigest(points, digest)
	if err != nil {
		log.Fatal(err.Error())
	}

	err = common.UploadDigestToS3(upl, digest)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Infof("Bootstrapping complete")
}
