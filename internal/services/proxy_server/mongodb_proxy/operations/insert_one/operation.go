package insert_one

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"log/slog"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/objects"
)

type Operation struct {
	collection  *mongo.Collection
	uuid        string
	isFinished  bool
	result      *mongo.InsertOneResult
	resultError error
}

func New(collection *mongo.Collection) *Operation {
	return &Operation{
		collection: collection,
		uuid:       uuid.New().String(),
		isFinished: false,
	}
}

func (a *Operation) IsFinished() bool {
	return a.isFinished
}

func (a *Operation) Uuid() string {
	return a.uuid
}

func (a *Operation) Start(ctx context.Context, document interface{}) *objects.RunningOperation {
	go a.execute(ctx, document)

	return &objects.RunningOperation{Uuid: a.uuid}
}

func (a *Operation) Result() (*mongo.InsertOneResult, error) {
	return a.result, a.resultError
}

func (a *Operation) execute(ctx context.Context, document interface{}) {
	slog.Info("InsertOne executing: " + a.uuid)

	a.result, a.resultError = a.collection.InsertOne(ctx, document)

	a.isFinished = true

	if a.resultError != nil {
		slog.Error("InsertOne error: " + a.uuid + " " + a.resultError.Error())
	} else {
		slog.Info("InsertOne success: " + a.uuid + " " + fmt.Sprintf("%s", a.result.InsertedID))
	}
}
