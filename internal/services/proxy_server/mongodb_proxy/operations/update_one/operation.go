package update_one

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
)

type Operation struct {
	filter      interface{}
	update      interface{}
	isFinished  bool
	result      *mongo.UpdateResult
	resultError error
}

func New(filter interface{}, update interface{}) *Operation {
	return &Operation{
		filter:     filter,
		update:     update,
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
	o.result, o.resultError = coll.UpdateOne(ctx, o.filter, o.update)

	o.isFinished = true
}

func (o *Operation) Result() (*mongo.UpdateResult, error) {
	return o.result, o.resultError
}
