package update_one

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Operation struct {
	filter      interface{}
	update      interface{}
	opUpsert    bool
	isFinished  bool
	result      *mongo.UpdateResult
	resultError error
}

type Options struct {
	Upsert bool
}

func New(
	filter interface{},
	update interface{},
	opUpsert bool,
) *Operation {
	return &Operation{
		filter:     filter,
		update:     update,
		opUpsert:   opUpsert,
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
	o.result, o.resultError = coll.UpdateOne(
		ctx,
		o.filter,
		o.update,
		&options.UpdateOptions{
			Upsert: &o.opUpsert,
		},
	)

	o.isFinished = true
}

func (o *Operation) Result() (interface{}, error) {
	return o.result, o.resultError
}
