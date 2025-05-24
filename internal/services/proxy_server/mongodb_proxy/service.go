package mongodb_proxy

import (
	"context"
	"log/slog"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/connections"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/insert_one"
)

type Service struct {
	ctx           context.Context
	connections   *connections.Connections
	insertOneList *insert_one.Operations
}

func NewService(ctx context.Context) *Service {
	slog.Info("Starting mongodb-proxy service...")

	return &Service{
		ctx:           ctx,
		connections:   connections.NewConnections(ctx),
		insertOneList: insert_one.NewOperations(),
	}
}

func (s *Service) InsertOne(
	connection string,
	database string,
	collection string,
	document interface{},
) (*objects.RunningOperation, error) {
	slog.Info("InsertOne: " + connection + " " + database + " " + collection)

	coll, err := s.connections.Collection(connection, database, collection)

	if err != nil {
		return nil, err
	}

	action := insert_one.New(coll)

	s.insertOneList.Add(action)

	return action.Start(s.ctx, document), nil
}

func (s *Service) InsertOneResult(operationUuid string) *insert_one.Operation {
	return s.insertOneList.Pull(operationUuid)
}

func (s *Service) Close() error {
	slog.Warn("Closing mongodb-proxy service")

	return s.connections.Close()
}
