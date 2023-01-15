package main

import (
	"bytes"
	"github.com/aws/aws-lambda-go/lambda"
	"strings"
	"time"

	"github.com/andrewfrench/locations/pkg/common"

	"github.com/apex/log"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/andrewfrench/ghcomp/pkg/ghcomp"
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

		digest, err := common.LoadDigestFromS3(dwn)
		if err != nil {
			panic(err)
		}

		tree := ghcomp.New(common.GeohashPrecision)
		buf := bytes.NewBuffer(nil)
		_, err = buf.WriteString(strings.Join(digest.Points, "\n"))
		if err != nil {
			log.Fatalf("failed to write string to buffer: %v", err)
		}

		window := []byte(digest.Points[0])
		for _, v := range digest.Points {
			offset := len(window) - len(v)
			for j := 0; j < len(v); j++ {
				window[offset+j] = v[j]
			}

			err = tree.Entree(window)
			if err != nil {
				log.Fatalf("failed to entree: %v", err)
			}
		}

		first, last, err := common.GetPointsBetween(dyn, tree, time.Now().Add(time.Duration(-24)*time.Hour), time.Now())
		if err != nil {
			return err
		}

		// Update digest metadata
		digest.GeneratedAt = int(time.Now().Unix())

		if int(first.Unix()) < digest.FirstTimestamp {
			digest.FirstTimestamp = int(first.Unix())
		}

		if int(last.Unix()) > digest.LastTimestamp {
			digest.LastTimestamp = int(last.Unix())
		}

		digest.Points = make([]string, 0)
		err = tree.WriteDeflated(digest)
		if err != nil {
			log.Fatalf("failed to write deflated geohashes to digest: %v", err)
		}

		digest.Size = len(digest.Points)

		err = common.UploadDigestToS3(upl, digest)
		if err != nil {
			return err
		}

		log.Infof("Reduction complete")

		return nil
	})
}
