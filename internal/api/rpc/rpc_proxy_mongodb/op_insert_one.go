package rpc_proxy_mongodb

type InsertOneArgs struct {
	Connection string
	Database   string
	Collection string
	Document   interface{}
}

type InsertOneReply struct {
	Error      string
	ActionUuid string
}

type InsertOneResultArgs struct {
	ActionUuid string
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
	operation, err := p.service.InsertOne(
		args.Connection,
		args.Database,
		args.Collection,
		args.Document,
	)

	if err != nil {
		reply.Error = err.Error()
	} else {
		reply.ActionUuid = operation.Uuid
	}

	return nil
}

func (p *ProxyMongodbServer) InsertOneResult(args *InsertOneResultArgs, reply *InsertOneResultReply) error {
	operation := p.service.InsertOneResult(args.ActionUuid)

	if !operation.IsFinished() {
		reply.IsFinished = false
	} else {
		reply.IsFinished = true

		result, resultError := operation.Result()

		if resultError != nil {
			reply.Error = resultError.Error()
		} else {
			reply.Result = InsertOneResult{
				InsertedID: result.InsertedID,
			}
		}
	}

	return nil
}
