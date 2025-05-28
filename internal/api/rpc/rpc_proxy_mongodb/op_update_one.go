package rpc_proxy_mongodb

import (
	"log/slog"
)

type UpdateOneArgs struct {
	Connection string
	Database   string
	Collection string
	Filter     string
	Update     string
	OpUpsert   bool
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
		msg := "unmarshal [Filter] json error: " + err.Error()

		reply.Error = msg

		slog.Error("UpdateOne: " + msg)

		return nil
	}

	update, err := unmarshalJson(args.Update)

	if err != nil {
		msg := "unmarshal [Update] json error: " + err.Error()

		reply.Error = msg

		slog.Error("UpdateOne: " + msg)

		return nil
	}

	runningOperation := p.service.UpdateOne(
		args.Connection,
		args.Database,
		args.Collection,
		filter,
		update,
		args.OpUpsert,
	)

	slog.Debug("UpdateOne[" + runningOperation.Uuid + "]: created")

	reply.OperationUuid = runningOperation.Uuid

	return nil
}

func (p *ProxyMongodbServer) UpdateOneResult(args *ResultArgs, reply *ResultReply) error {
	operation := p.service.UpdateOneResult(args.OperationUuid)

	p.makeResult(operation, reply)

	return nil
}
