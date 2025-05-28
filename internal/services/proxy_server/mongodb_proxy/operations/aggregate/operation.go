package aggregate

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
)

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
		pipeline:   o.pipeline,
		isFinished: false,
		cursor:     o.cursor,
	}

	go cloned.next(ctx)

	return cloned
}

func (o *Operation) next(ctx context.Context) {
	defer func(o *Operation) {
		o.isFinished = true
	}(o)

	if !o.cursor.Next(ctx) {
		slog.Debug("Aggregate 1")

		o.result = nil

		_ = o.cursor.Close(ctx)

		return
	}

	if err := o.cursor.Err(); err != nil {
		slog.Debug("Aggregate 2")

		o.result = nil
		o.resultError = err

		_ = o.cursor.Close(ctx)

		return
	}

	slog.Debug("Aggregate 3")

	o.result = o.cursor.Current
}
