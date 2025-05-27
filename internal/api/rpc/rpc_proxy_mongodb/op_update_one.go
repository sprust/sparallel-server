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

func (p *ProxyMongodbServer) UpdateOneResult(args *ResultArgs, reply *ResultReply) error {
	operation := p.service.UpdateOneResult(args.OperationUuid)

	p.makeResult(operation, reply)

	return nil
}
