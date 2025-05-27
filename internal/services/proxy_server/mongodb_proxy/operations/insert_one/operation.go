package insert_one

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
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

func (o *Operation) Result() (interface{}, error) {
	return o.result, o.resultError
}
