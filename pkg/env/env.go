package env

import (
	"fmt"
	"os"
	"strconv"
)

const (
	AccuracyLimitVar   = "ACCURACY_LIMIT"
	DistributionIdVar  = "DISTRIBUTION_ID"
	LocationsBucketVar = "LOCATIONS_BUCKET"
	LocationsTableVar  = "LOCATIONS_TABLE"
	PartialsObjectVar  = "PARTIALS_OBJECT"
	IndexNameVar       = "INDEX_NAME"
	TrackerIDVar       = "TRACKER_ID"
	PartialsKeyVar     = "PARTIALS_KEY"
)

func AccuracyLimit() (int, error) {
	strVal, err := requiredVar(AccuracyLimitVar)
	if err != nil {
		return 0, err
	}

	intVal, err := strconv.Atoi(strVal)
	if err != nil {
		return 0, fmt.Errorf("failed to parse %s: %v", AccuracyLimitVar, err)
	}

	return intVal, nil
}

func LocationsTable() (string, error) {
	return requiredVar(LocationsTableVar)
}

func LocationsBucket() (string, error) {
	return requiredVar(LocationsBucketVar)
}

func DistributionId() (string, error) {
	return requiredVar(DistributionIdVar)
}

func PartialsObject() (string, error) {
	return requiredVar(PartialsObjectVar)
}

func IndexName() (string, error) {
	return requiredVar(IndexNameVar)
}

func TrackerID() (string, error) {
	return requiredVar(TrackerIDVar)
}

func PartialsKey() (string, error) {
	return requiredVar(PartialsKeyVar)
}

func requiredVar(varName string) (string, error) {
	strVal := os.Getenv(varName)
	if strVal == "" {
		return "", fmt.Errorf("%s env var is required", varName)
	}

	return strVal, nil
}
