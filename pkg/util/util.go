package util

import (
	"crypto/md5"
	"fmt"

	"github.com/andrewfrench/owntracks-go/pkg/owntracks"
)

func GenerateId(l *owntracks.Location) (string, error) {
	hashInput := fmt.Sprintf("%s%d%d", l.Topic, l.ReportedTimestamp, l.ReceivedTimestamp)
	hash := md5.Sum([]byte(hashInput))

	return fmt.Sprintf("%x", hash), nil
}
