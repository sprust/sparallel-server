package rpc_proxy_mongodb

import "log/slog"

type BulkWriteArgs struct {
	Connection string
	Database   string
	Collection string
	Models     string
}

type BulkWriteReply struct {
	Error         string
	OperationUuid string
}

func (p *ProxyMongodbServer) BulkWrite(args *BulkWriteArgs, reply *BulkWriteReply) error {
	models, err := unmarshalModels(args.Models)

	if err != nil {
		msg := "unmarshal [Models] json error: " + err.Error()

		slog.Error("BulkWrite: " + msg)

		reply.Error = msg

		return nil
	}

	runningOperation := p.service.BulkWrite(
		args.Connection,
		args.Database,
		args.Collection,
		models,
	)

	slog.Debug("BulkWrite[" + runningOperation.Uuid + "]: created")

	reply.OperationUuid = runningOperation.Uuid

	return nil
}

func (p *ProxyMongodbServer) BulkWriteResult(args *ResultArgs, reply *ResultReply) error {
	operation := p.service.BulkWriteResult(args.OperationUuid)

	p.makeResult(operation, "", reply)

	return nil
}
