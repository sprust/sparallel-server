package insert_one

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
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
	a.uuid = uuid.New().String()

	go a.execute(ctx, document)

	return &objects.RunningOperation{Uuid: a.uuid}
}

func (a *Operation) Result() (*mongo.InsertOneResult, error) {
	return a.result, a.resultError
}

func (a *Operation) execute(ctx context.Context, document interface{}) {
	a.result, a.resultError = a.collection.InsertOne(ctx, document)

	a.isFinished = true
}
