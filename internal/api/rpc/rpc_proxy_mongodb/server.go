package rpc_proxy_mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"log/slog"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
	"sync"
)

var server *ProxyMongodbServer
var once sync.Once

type ProxyMongodbServer struct {
	ctx     context.Context
	service *mongodb_proxy.Service
}

type ResultArgs struct {
	OperationUuid string
}

type ResultReply struct {
	IsFinished bool
	Error      string
	Result     string
	NextUuid   string
}

func NewServer(ctx context.Context) *ProxyMongodbServer {
	once.Do(func() {
		server = &ProxyMongodbServer{
			ctx:     ctx,
			service: mongodb_proxy.NewService(ctx),
		}
	})

	return server
}

func (p *ProxyMongodbServer) makeResult(
	operation operations.OperationInterface,
	nextOpUuid string,
	reply *ResultReply,
) {
	if operation == nil {
		reply.IsFinished = false
	} else {
		reply.IsFinished = operation.IsFinished()

		if reply.IsFinished {
			result, resultError := operation.Result()

			if resultError != nil {
				reply.Error = resultError.Error()
			} else {
				if result == nil {
					result = bson.D{}
				}

				serialized, err := serializeForPHPMongoDB(result)

				if err != nil {
					reply.Error = err.Error()
				} else {
					reply.Result = serialized
					reply.NextUuid = nextOpUuid
				}
			}
		}
	}
}

//
//func (s *ProxyMongodbServer) InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
//	return c.collection.InsertMany(ctx, documents)
//}
//
//func (s *ProxyMongodbServer) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
//	return c.collection.UpdateOne(ctx, filter, update, opts...)
//}
//
//func (s *ProxyMongodbServer) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
//	return c.collection.UpdateMany(ctx, filter, update, opts...)
//}
//
//func (s *ProxyMongodbServer) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
//	return c.collection.DeleteOne(ctx, filter, opts...)
//}
//
//func (s *ProxyMongodbServer) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
//	return c.collection.DeleteMany(ctx, filter, opts...)
//}
//
//func (s *ProxyMongodbServer) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
//	return c.collection.Find(ctx, filter, opts...)
//}
//
//func (s *ProxyMongodbServer) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
//	return c.collection.FindOne(ctx, filter, opts...)
//}
//
//func (s *ProxyMongodbServer) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
//	return c.collection.CountDocuments(ctx, filter, opts...)
//}
//
//func (s *ProxyMongodbServer) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
//	return c.collection.Aggregate(ctx, pipeline, opts...)
//}
//
//func (s *ProxyMongodbServer) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
//	return c.collection.BulkWrite(ctx, models, opts...)
//}
//
//func (s *ProxyMongodbServer) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
//	return c.collection.Watch(ctx, pipeline, opts...)
//}
//
//func (s *ProxyMongodbServer) Drop(ctx context.Context) error {
//	return c.collection.Drop(ctx)
//}
//
//func (s *ProxyMongodbServer) CreateIndex(ctx context.Context, model mongo.IndexModel) (string, error) {
//	return c.collection.Indexes().CreateOne(ctx, model)
//}
//
//func (s *ProxyMongodbServer) CreateIndexes(ctx context.Context, models []mongo.IndexModel) ([]string, error) {
//	return c.collection.Indexes().CreateMany(ctx, models)
//}

func (p *ProxyMongodbServer) Pause() error {
	return nil
}

func (p *ProxyMongodbServer) UnPause() error {
	return nil
}

func (p *ProxyMongodbServer) Close() error {
	slog.Warn("Closing mongodb-proxy server")

	return p.service.Close()
}
