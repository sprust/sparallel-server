package bulk_write

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
)

type Operation struct {
	models      []mongo.WriteModel
	isFinished  bool
	result      *mongo.BulkWriteResult
	resultError error
}

func New(models []mongo.WriteModel) *Operation {
	return &Operation{
		models:     models,
		isFinished: false,
	}
}

func (o *Operation) IsFinished() bool {
	return o.isFinished
}

func (o *Operation) Error(err error) {
	o.isFinished = true
	o.resultError = err
}

func (o *Operation) Execute(ctx context.Context, coll *mongo.Collection) {
	o.result, o.resultError = coll.BulkWrite(ctx, o.models)

	o.isFinished = true
}

func (o *Operation) HasNext() bool {
	return false
}

func (o *Operation) Clone(context.Context) operations.OperationInterface {
	panic("forbidden")
}

func (o *Operation) Result() (interface{}, error) {
	return o.result, o.resultError
}
