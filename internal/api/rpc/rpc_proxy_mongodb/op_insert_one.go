package rpc_proxy_mongodb

import "log/slog"

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
		msg := "unmarshal [Document] json error: " + err.Error()

		slog.Error("InsertOne: " + msg)

		reply.Error = msg

		return nil
	}

	runningOperation := p.service.InsertOne(
		args.Connection,
		args.Database,
		args.Collection,
		document,
	)

	slog.Debug("InsertOne[" + runningOperation.Uuid + "]: created")

	reply.OperationUuid = runningOperation.Uuid

	return nil
}

func (p *ProxyMongodbServer) InsertOneResult(args *ResultArgs, reply *ResultReply) error {
	operation := p.service.InsertOneResult(args.OperationUuid)

	p.makeResult(operation, reply)

	return nil
}
