package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/andrewfrench/ghcomp/pkg/ghcomp"
	"io"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/apex/log"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	Table             = "locations"
	Bucket            = "afrench-locations"
	DigestKey         = "data/digest.json"
	ReportedTimestamp = "reportedTimestamp"
	GeoHash           = "geoHash"
	Region            = "us-west-2"
	ContentType       = "application/json"
	QueryLimit        = 1024
	AccuracyLimit     = 100
	Index             = "trackerId-reportedTimestamp-index"
	GeohashPrecision  = 9
)
var trackerIDs = []string{"8E", "AD", "FB", "3E"}

type Digest struct {
	Size           int      `json:"size"`
	FirstTimestamp int      `json:"firstTimestamp"`
	LastTimestamp  int      `json:"lastTimestamp"`
	GeneratedAt    int      `json:"generatedAt"`
	Points         []string `json:"points"`
	readIndex      int
}

// Write satisfies io.Writer by appending the incoming byte slice to the Points slice byte by byte.
// This method does not update digest metadata; this is left to the caller to update when point
// data has been fully written to the digest.
func (d *Digest) Write(p []byte) (int, error) {
	if len(d.Points) == 0 {
		d.Points = append(d.Points, "")
	}

	if string(p) == "\n" {
		d.Points = append(d.Points, "")
	} else {
		d.Points[len(d.Points)-1] = d.Points[len(d.Points)-1] + string(p)
	}

	return len(p), nil
}

// Read satisfies io.Reader by copying the values of digest.Points, in order, to p for each successive call.
// This method returns io.EOF when digest.Points has been exhausted.
func (d *Digest) Read(p []byte) (int, error) {
	if d.readIndex >= len(d.Points) {
		return 0, io.EOF
	}

	nextValue := d.Points[d.readIndex] + "\n"

	// If the remaining length of the input buffer is less than the length of the next value, fill
	// the buffer and return.
	if len(p) < len(nextValue) {
		for i := range p {
			p[i] = nextValue[i]
		}

		return len(p), nil
	}

	// If the remaining length of the buffer exceeds the length of the next value, add it to the buffer.
	for i := range nextValue {
		p[i] = nextValue[i]
	}

	d.readIndex++

	fmt.Print(nextValue)
	return len(nextValue), nil
}

func LoadDigestFromS3(dwn *s3manager.Downloader) (*Digest, error) {
	digest := new(Digest)

	log.Infof("Downloading %s/%s", Bucket, DigestKey)
	buf := aws.NewWriteAtBuffer([]byte{})
	_, err := dwn.Download(buf, &s3.GetObjectInput{
		Bucket: aws.String(Bucket),
		Key:    aws.String(DigestKey),
	})
	if err != nil {
		return nil, err
	}

	log.Infof("Unmarshalling data to digest struct")
	err = json.Unmarshal(buf.Bytes(), digest)
	if err != nil {
		return nil, err
	}

	log.Infof("Loaded %d points from existing digest", len(digest.Points))

	return digest, nil
}

func GetPointsBetween(dyn *dynamodb.DynamoDB, tree *ghcomp.Tree, from time.Time, to time.Time) (time.Time, time.Time, error) {
	log.Infof("Loading data from DynamoDB %s/%s", Table, Index)
	log.Infof("Query size: %d", QueryLimit)
	log.Infof("Query from: %s", from.String())
	log.Infof("Query to: %s", to.String())

	first := to
	last := from
	for _, trackerID := range trackerIDs {
		var next map[string]*dynamodb.AttributeValue = nil
		for {
			if next == nil {
				log.Infof("Querying from %s", from.String())
			} else {
				ts, _ := strconv.Atoi(*next[ReportedTimestamp].N)
				log.Infof("Querying from %s (%d)", time.Unix(int64(ts), 0).String(), ts)
			}

			output, err := dyn.Query(&dynamodb.QueryInput{
				TableName:         aws.String(Table),
				ExclusiveStartKey: next,
				AttributesToGet: []*string{
					aws.String("geoHash"),
					aws.String("reportedTimestamp"),
					aws.String("accuracy"),
				},
				Limit:     aws.Int64(QueryLimit),
				IndexName: aws.String(Index),
				KeyConditions: map[string]*dynamodb.Condition{
					"trackerId": {
						ComparisonOperator: aws.String("EQ"),
						AttributeValueList: []*dynamodb.AttributeValue{
							{
								S: &trackerID,
							},
						},
					},
					"reportedTimestamp": {
						ComparisonOperator: aws.String("BETWEEN"),
						AttributeValueList: []*dynamodb.AttributeValue{
							{
								N: aws.String(fmt.Sprintf("%d", from.Unix())),
							},
							{
								N: aws.String(fmt.Sprintf("%d", to.Unix())),
							},
						},
					},
				},
			})
			if err != nil {
				return first, last, err
			}

			accuracyRejected := 0
			for _, item := range output.Items {
				accuracyValue, err := strconv.Atoi(*item["accuracy"].N)
				if err != nil {
					return first, last, err
				}

				if accuracyValue > AccuracyLimit {
					accuracyRejected++
					continue
				}

				timeInt, err := strconv.Atoi(*item[ReportedTimestamp].N)
				if err != nil {
					return first, last, fmt.Errorf("failed to parse reported timestamp: %v", err)
				}

				t := time.Unix(int64(timeInt), 0)
				if t.Before(first) {
					first = t
				}

				if t.After(last) {
					last = t
				}

				err = tree.Entree([]byte(*item[GeoHash].S)[:GeohashPrecision])
				if err != nil {
					return first, last, err
				}
			}

			next = output.LastEvaluatedKey
			if next == nil {
				break
			}
		}
	}

	return first, last, nil
}

//func BuildDigest(points []string) error {
//	log.Infof("Building digest")
//
//	digest := new(Digest)
//	digest.Size = len(points)
//	digest.GeneratedAt = int(time.Now().Unix())
//	for k, v := range points {
//		t, err := strconv.Atoi(k)
//		if err != nil {
//			return err
//		}
//
//		if t < digest.FirstTimestamp {
//			digest.FirstTimestamp = t
//		}
//
//		if t > digest.LastTimestamp {
//			digest.LastTimestamp = t
//		}
//
//		digest.Points = append(digest.Points, &Point{
//			Lat: fmt.Sprintf("%0.4f", lat),
//			Lng: fmt.Sprintf("%0.4f", lng),
//		})
//	}
//	log.Info("Built digest")
//
//	return nil
//}

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
