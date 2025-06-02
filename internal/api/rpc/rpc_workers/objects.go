package rpc_workers

type AddTaskArgs struct {
	GroupUuid   string
	TaskUuid    string
	UnixTimeout int
	Payload     string
}

type AddTaskResult struct {
	Uuid string
}

type DetectFinishedTaskArgs struct {
	GroupUuid string
}

type DetectFinishedTaskResult struct {
	GroupUuid  string
	TaskUuid   string
	IsFinished bool
	Response   string
	IsError    bool
}

type CancelGroupArgs struct {
	GroupUuid string
}

type CancelGroupResult struct {
	GroupUuid string
}
