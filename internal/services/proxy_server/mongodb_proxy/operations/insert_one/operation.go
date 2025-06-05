package insert_one

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
)

type Operation struct {
	document    interface{}
	isFinished  bool
	result      *mongo.InsertOneResult
	resultError error
}

func New(document interface{}) *Operation {
	return &Operation{
		document:   document,
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
	o.result, o.resultError = coll.InsertOne(ctx, o.document)

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
