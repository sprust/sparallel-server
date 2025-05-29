package mongodb_proxy

import (
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/mongodb_proxy_objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/aggregate"
)

func (s *Service) Aggregate(
	connection string,
	database string,
	collection string,
	pipeline interface{},
) *mongodb_proxy_objects.RunningOperation {
	return s.aggregateList.Add(
		s.ctx,
		connection,
		database,
		collection,
		aggregate.New(pipeline),
	)
}

func (s *Service) AggregateResult(operationUuid string) (operations.OperationInterface, string) {
	op, nextOpUuid := s.aggregateList.Pull(s.ctx, operationUuid)

	return op, nextOpUuid
}
