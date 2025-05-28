package rpc_proxy_mongodb

import "log/slog"

type AggregateArgs struct {
	Connection string
	Database   string
	Collection string
	Pipeline   string
}

type AggregateReply struct {
	Error         string
	OperationUuid string
}

func (p *ProxyMongodbServer) Aggregate(args *AggregateArgs, reply *AggregateReply) error {
	pipeline, err := unmarshalJson(args.Pipeline)

	if err != nil {
		msg := "unmarshal [Pipeline] json error: " + err.Error()

		slog.Error("Aggregate: " + msg)

		reply.Error = msg

		return nil
	}

	runningOperation := p.service.Aggregate(
		args.Connection,
		args.Database,
		args.Collection,
		pipeline,
	)

	slog.Debug("Aggregate[" + runningOperation.Uuid + "]: created")

	reply.OperationUuid = runningOperation.Uuid

	return nil
}

func (p *ProxyMongodbServer) AggregateResult(args *ResultArgs, reply *ResultReply) error {
	operation, nextUuid := p.service.AggregateResult(args.OperationUuid)

	p.makeResult(operation, nextUuid, reply)

	return nil
}
