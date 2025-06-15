package aggregate

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
)

const resultKey = "_result"

type Operation struct {
	pipeline    interface{}
	isFinished  bool
	cursor      *mongo.Cursor
	result      interface{}
	resultError error
}

func New(pipeline interface{}) *Operation {
	return &Operation{
		pipeline:   pipeline,
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
	o.cursor, o.resultError = coll.Aggregate(ctx, o.pipeline)

	if o.resultError != nil {
		o.isFinished = true

		return
	}

	o.next(ctx)
}

func (o *Operation) Result() (interface{}, error) {
	return o.result, o.resultError
}

func (o *Operation) HasNext() bool {
	return o.isFinished && o.result != nil
}

func (o *Operation) Clone(ctx context.Context) operations.OperationInterface {
	cloned := &Operation{
		isFinished: false,
		cursor:     o.cursor,
	}

	cloned.next(ctx)

	return cloned
}

func (o *Operation) next(ctx context.Context) {
	defer func(o *Operation) {
		o.isFinished = true
	}(o)

	counter := 20

	var items []interface{}

	for o.cursor.Next(ctx) && counter > 0 {
		if err := o.cursor.Err(); err != nil {
			o.result = nil
			o.resultError = err

			_ = o.cursor.Close(ctx)

			break
		}

		items = append(items, o.cursor.Current)

		counter -= 1
	}

	if len(items) == 0 {
		_ = o.cursor.Close(ctx)
	} else {
		o.result = bson.D{
			struct {
				Key   string
				Value interface{}
			}{
				Key:   resultKey,
				Value: items,
			},
		}
	}
}
