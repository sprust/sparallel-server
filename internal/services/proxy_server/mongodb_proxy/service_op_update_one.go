package mongodb_proxy

import (
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/mongodb_proxy_objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/update_one"
)

func (s *Service) UpdateOne(
	connection string,
	database string,
	collection string,
	filter interface{},
	update interface{},
	opUpsert bool,
) *mongodb_proxy_objects.RunningOperation {
	return s.updateOneList.Add(
		s.ctx,
		connection,
		database,
		collection,
		update_one.New(
			filter,
			update,
			opUpsert,
		),
	)
}

func (s *Service) UpdateOneResult(operationUuid string) operations.OperationInterface {
	op, _ := s.updateOneList.Pull(s.ctx, operationUuid)

	return op
}
