package main

import (
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
	cfg := &aws.Config{Region: aws.String(common.Region)}
	sess, err := session.NewSession(cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	dyn := dynamodb.New(sess)
	upl := s3manager.NewUploader(sess)

	// Bootstrapping requires pulling all points, start from time 0.
	tree := ghcomp.New(common.GeohashPrecision)
	first, last, err := common.GetPointsBetween(dyn, tree, time.Unix(0, 0), time.Now())
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

	err = common.UploadDigestToS3(upl, digest)
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Infof("Bootstrapping complete")
}
