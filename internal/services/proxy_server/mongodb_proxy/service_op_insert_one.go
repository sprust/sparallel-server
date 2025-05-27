package mongodb_proxy

import (
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/insert_one"
)

func (s *Service) InsertOne(
	connection string,
	database string,
	collection string,
	document interface{},
) *objects.RunningOperation {
	return s.insertOneList.Add(
		s.ctx,
		connection,
		database,
		collection,
		insert_one.New(document),
	)
}

func (s *Service) InsertOneResult(operationUuid string) *insert_one.Operation {
	op := s.insertOneList.Pull(operationUuid)

	if op == nil {
		return nil
	}

	return op.(*insert_one.Operation)
}
