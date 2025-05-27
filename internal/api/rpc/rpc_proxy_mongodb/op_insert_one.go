package rpc_proxy_mongodb

type InsertOneArgs struct {
	Connection string
	Database   string
	Collection string
	Document   string
}

type InsertOneReply struct {
	Error         string
	OperationUuid string
}

func (p *ProxyMongodbServer) InsertOne(args *InsertOneArgs, reply *InsertOneReply) error {
	document, err := unmarshalJson(args.Document)

	if err != nil {
		reply.Error = err.Error()

		return nil
	}

	runningOperation := p.service.InsertOne(
		args.Connection,
		args.Database,
		args.Collection,
		document,
	)

	reply.OperationUuid = runningOperation.Uuid

	return nil
}

func (p *ProxyMongodbServer) InsertOneResult(args *ResultArgs, reply *ResultReply) error {
	operation := p.service.InsertOneResult(args.OperationUuid)

	p.makeResult(operation, reply)

	return nil
}
