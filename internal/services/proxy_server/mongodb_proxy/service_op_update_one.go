package mongodb_proxy

import (
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/update_one"
)

func (s *Service) UpdateOne(
	connection string,
	database string,
	collection string,
	filter interface{},
	update interface{},
	opUpsert bool,
) *objects.RunningOperation {
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

func (s *Service) UpdateOneResult(operationUuid string) *update_one.Operation {
	op := s.updateOneList.Pull(operationUuid)

	if op == nil {
		return nil
	}

	return op.(*update_one.Operation)
}
