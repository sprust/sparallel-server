package operations

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type OperationInterface interface {
	IsFinished() bool
	Error(err error)
	Execute(ctx context.Context, coll *mongo.Collection)
	HasNext() bool
	Clone(ctx context.Context) OperationInterface
	Result() (interface{}, error)
}
