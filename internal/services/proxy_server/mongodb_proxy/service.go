package mongodb_proxy

import (
	"context"
	"log/slog"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/connections"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/mongodb_proxy_objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/aggregate"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/insert_one"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/update_one"
	"sync"
)

var service *Service
var once sync.Once

type Service struct {
	ctx           context.Context
	connections   *connections.Connections
	insertOneList *operations.Operations[*insert_one.Operation]
	updateOneList *operations.Operations[*update_one.Operation]
	aggregateList *operations.Operations[*aggregate.Operation]
}

func NewService(ctx context.Context) *Service {
	slog.Info("Starting mongodb-proxy service...")

	once.Do(func() {
		connFactory := connections.NewConnections(ctx)

		service = &Service{
			ctx:         ctx,
			connections: connFactory,
			insertOneList: operations.NewOperations[*insert_one.Operation](
				ctx,
				"InsertOne",
				connFactory,
			),
			updateOneList: operations.NewOperations[*update_one.Operation](
				ctx,
				"UpdateOne",
				connFactory,
			),
			aggregateList: operations.NewOperations[*aggregate.Operation](
				ctx,
				"Aggregate",
				connFactory,
			),
		}
	})

	return service
}

func GetService() *Service {
	return service
}

func (s *Service) Stats() mongodb_proxy_objects.ServiceStats {
	return mongodb_proxy_objects.ServiceStats{
		InsertOne: s.insertOneList.Stats(),
		UpdateOne: s.updateOneList.Stats(),
		Aggregate: s.aggregateList.Stats(),
	}
}

func (s *Service) Close() error {
	slog.Warn("Closing mongodb-proxy service")

	return s.connections.Close()
}
