package proxy_mongodb

import (
	"context"
	"log/slog"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy"
)

var server *Server

type Server struct {
	ctx     context.Context
	service *mongodb_proxy.Service
}

func NewMongodbProxyServer(ctx context.Context) *Server {
	if server != nil {
		panic("server already initialized")
	}

	return &Server{
		ctx:     ctx,
		service: mongodb_proxy.NewService(ctx),
	}
}

//
//func (s *Server) InsertMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
//	return c.collection.InsertMany(ctx, documents)
//}
//
//func (s *Server) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
//	return c.collection.UpdateOne(ctx, filter, update, opts...)
//}
//
//func (s *Server) UpdateMany(ctx context.Context, filter interface{}, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
//	return c.collection.UpdateMany(ctx, filter, update, opts...)
//}
//
//func (s *Server) DeleteOne(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
//	return c.collection.DeleteOne(ctx, filter, opts...)
//}
//
//func (s *Server) DeleteMany(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
//	return c.collection.DeleteMany(ctx, filter, opts...)
//}
//
//func (s *Server) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (*mongo.Cursor, error) {
//	return c.collection.Find(ctx, filter, opts...)
//}
//
//func (s *Server) FindOne(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
//	return c.collection.FindOne(ctx, filter, opts...)
//}
//
//func (s *Server) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
//	return c.collection.CountDocuments(ctx, filter, opts...)
//}
//
//func (s *Server) Aggregate(ctx context.Context, pipeline interface{}, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
//	return c.collection.Aggregate(ctx, pipeline, opts...)
//}
//
//func (s *Server) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
//	return c.collection.BulkWrite(ctx, models, opts...)
//}
//
//func (s *Server) Watch(ctx context.Context, pipeline interface{}, opts ...*options.ChangeStreamOptions) (*mongo.ChangeStream, error) {
//	return c.collection.Watch(ctx, pipeline, opts...)
//}
//
//func (s *Server) Drop(ctx context.Context) error {
//	return c.collection.Drop(ctx)
//}
//
//func (s *Server) CreateIndex(ctx context.Context, model mongo.IndexModel) (string, error) {
//	return c.collection.Indexes().CreateOne(ctx, model)
//}
//
//func (s *Server) CreateIndexes(ctx context.Context, models []mongo.IndexModel) ([]string, error) {
//	return c.collection.Indexes().CreateMany(ctx, models)
//}

func (p *Server) Close() error {
	slog.Warn("Closing mongodb-proxy server")

	return p.service.Close()
}
