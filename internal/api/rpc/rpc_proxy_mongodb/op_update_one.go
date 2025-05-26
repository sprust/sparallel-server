package rpc_proxy_mongodb

type UpdateOneArgs struct {
	Connection string
	Database   string
	Collection string
	Filter     string
	Update     string
}

type UpdateOneReply struct {
	Error         string
	OperationUuid string
}

type UpdateOneResultArgs struct {
	OperationUuid string
}

type UpdateOneResultReply struct {
	IsFinished bool
	Error      string
	Result     UpdateOneResult
}

type UpdateOneResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    interface{}
}

func (p *ProxyMongodbServer) UpdateOne(args *UpdateOneArgs, reply *UpdateOneReply) error {
	filter, err := unmarshalJson(args.Filter)

	if err != nil {
		reply.Error = err.Error()

		return nil
	}

	update, err := unmarshalJson(args.Update)

	if err != nil {
		reply.Error = err.Error()

		return nil
	}

	runningOperation := p.service.UpdateOne(
		args.Connection,
		args.Database,
		args.Collection,
		filter,
		update,
	)

	reply.OperationUuid = runningOperation.Uuid

	return nil
}

func (p *ProxyMongodbServer) UpdateOneResult(args *UpdateOneResultArgs, reply *UpdateOneResultReply) error {
	operation := p.service.UpdateOneResult(args.OperationUuid)

	if operation == nil {
		reply.IsFinished = true
		reply.Error = "unexisting operation"
	} else {
		reply.IsFinished = operation.IsFinished()

		if reply.IsFinished {
			result, resultError := operation.Result()

			if resultError != nil {
				reply.Error = resultError.Error()
			} else {
				reply.Result = UpdateOneResult{
					MatchedCount:  result.MatchedCount,
					ModifiedCount: result.ModifiedCount,
					UpsertedCount: result.UpsertedCount,
					UpsertedID:    result.UpsertedID,
				}
			}
		}
	}

	return nil
}
