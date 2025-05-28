package mongodb_proxy

import (
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/objects"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/operations/aggregate"
)

func (s *Service) Aggregate(
	connection string,
	database string,
	collection string,
	pipeline interface{},
) *objects.RunningOperation {
	return s.aggregateList.Add(
		s.ctx,
		connection,
		database,
		collection,
		aggregate.New(pipeline),
	)
}

func (s *Service) AggregateNext() {

}

func (s *Service) AggregateResult(operationUuid string) (*aggregate.Operation, string) {
	op, nextUuid := s.aggregateList.Pull(s.ctx, operationUuid)

	if op == nil {
		return nil, ""
	}

	return op.(*aggregate.Operation), nextUuid
}
