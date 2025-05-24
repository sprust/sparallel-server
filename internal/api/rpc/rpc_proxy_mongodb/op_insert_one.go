package rpc_proxy_mongodb

import (
	"encoding/json"
)

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

type InsertOneResultArgs struct {
	OperationUuid string
}

type InsertOneResultReply struct {
	IsFinished bool
	Error      string
	Result     InsertOneResult
}

type InsertOneResult struct {
	InsertedID interface{}
}

func (p *ProxyMongodbServer) InsertOne(args *InsertOneArgs, reply *InsertOneReply) error {
	var document interface{}

	err := json.Unmarshal([]byte(args.Document), &document)

	if err != nil {
		reply.Error = err.Error()

		return nil
	}

	document = processDateValues(document)

	operation, err := p.service.InsertOne(
		args.Connection,
		args.Database,
		args.Collection,
		document,
	)

	if err != nil {
		reply.Error = err.Error()
	} else {
		reply.OperationUuid = operation.Uuid
	}

	return nil
}

func (p *ProxyMongodbServer) InsertOneResult(args *InsertOneResultArgs, reply *InsertOneResultReply) error {
	operation := p.service.InsertOneResult(args.OperationUuid)

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
				reply.Result = InsertOneResult{
					InsertedID: result.InsertedID,
				}
			}
		}
	}

	return nil
}
