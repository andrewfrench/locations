package main

import (
	"fmt"
	"time"

	"github.com/andrewfrench/locations/pkg/common"

	"github.com/apex/log"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func main() {
	lambda.Start(func() error {
		cfg := &aws.Config{Region: aws.String(common.Region)}
		sess, err := session.NewSession(cfg)
		if err != nil {
			return err
		}

		dyn := dynamodb.New(sess)
		dwn := s3manager.NewDownloader(sess)
		upl := s3manager.NewUploader(sess)

		// Points are stored as a map of reportedTimestamp -> geoHash
		points := map[string]string{}
		err = common.LoadCacheFromS3(dwn, &points)
		if err != nil {
			panic(err)
		}

		fromTime := fmt.Sprintf("%d", time.Now().Add(time.Duration(-24)*time.Hour).Unix())
		err = common.LoadRecentsFromDynamoDB(dyn, fromTime, &points)
		if err != nil {
			return err
		}

		err = common.UploadCacheToS3(upl, &points)
		if err != nil {
			return err
		}

		digest := &common.Digest{}
		err = common.BuildDigest(points, digest)
		if err != nil {
			return err
		}

		err = common.UploadDigestToS3(upl, digest)
		if err != nil {
			return err
		}

		log.Infof("Reduction complete")

		return nil
	})
}
