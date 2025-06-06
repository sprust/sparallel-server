package mongodb_proxy

import (
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/mongodb_proxy_objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/insert_one"
)

func (s *Service) InsertOne(
	connection string,
	database string,
	collection string,
	document interface{},
) *mongodb_proxy_objects.RunningOperation {
	return s.insertOneList.Add(
		s.ctx,
		connection,
		database,
		collection,
		insert_one.New(document),
	)
}

func (s *Service) InsertOneResult(operationUuid string) operations.OperationInterface {
	op, _ := s.insertOneList.Pull(s.ctx, operationUuid)

	return op
}
