package mongodb_proxy

import (
	"go.mongodb.org/mongo-driver/mongo"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/mongodb_proxy_objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/bulk_write"
)

func (s *Service) BulkWrite(
	connection string,
	database string,
	collection string,
	models []mongo.WriteModel,
) *mongodb_proxy_objects.RunningOperation {
	return s.bulkWriteList.Add(
		s.ctx,
		connection,
		database,
		collection,
		bulk_write.New(models),
	)
}

func (s *Service) BulkWriteResult(operationUuid string) operations.OperationInterface {
	op, _ := s.bulkWriteList.Pull(s.ctx, operationUuid)

	return op
}
