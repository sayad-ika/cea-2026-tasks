package repository

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const rfc3339 = time.RFC3339

var ErrConcurrentModification = errors.New("concurrent modification detected; retry")

func isConditionalCheckFailed(err error) bool {
	var ccf *types.ConditionalCheckFailedException
	return errors.As(err, &ccf)
}
