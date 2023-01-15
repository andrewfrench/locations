package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/andrewfrench/ghcomp/pkg/ghcomp"
	"github.com/andrewfrench/locations/pkg/common"
	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"time"
)

func main() {
	year := flag.Int("year", 2022, "Year to select points from (e.g. 2022).")
	flag.Parse()

	from, err := time.Parse(time.RFC3339, fmt.Sprintf("%d-01-01T00:00:00Z", *year))
	if err != nil {
		log.Fatalf("failed to parse date: %v", err)
	}

	to, err := time.Parse(time.RFC3339, fmt.Sprintf("%d-01-01T00:00:00Z", *year+1))
	if err != nil {
		log.Fatalf("failed to parse date: %v", err)
	}

	cfg := &aws.Config{Region: aws.String(common.Region)}
	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	dyn := dynamodb.New(sess)
	upl := s3manager.NewUploader(sess)

	// Bootstrapping requires pulling all points, start from time 0.
	tree := ghcomp.New(common.GeohashPrecision)
	first, last, err := common.GetPointsBetween(dyn, tree, from, to)
	if err != nil {
		log.Fatal(err.Error())
	}

	digest := new(common.Digest)
	err = tree.WriteDeflated(digest)
	if err != nil {
		log.Fatal(err.Error())
	}

	digest.Size = len(digest.Points)
	digest.FirstTimestamp = int(first.Unix())
	digest.LastTimestamp = int(last.Unix())
	digest.GeneratedAt = int(time.Now().Unix())

	log.Infof("Marshalling data")
	outb, err := json.Marshal(digest)
	if err != nil {
		log.Fatalf("failed to marshal digest: %v", err)
	}

	key := fmt.Sprintf("data/%d.json", *year)
	log.Infof("Uploading %s/%s", common.Bucket, key)
	_, err = upl.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(common.Bucket),
		Key:         aws.String(key),
		Body:        bytes.NewBuffer(outb),
		ContentType: aws.String(common.ContentType),
	})
	if err != nil {
		log.Fatalf("failed to upload: %v", err)
	}

	log.Infof("Upload complete")
}
