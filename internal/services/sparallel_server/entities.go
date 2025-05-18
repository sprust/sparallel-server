package sparallel_server

type Task struct {
	GroupUuid       string
	Uuid            string
	UnixTimeTimeout int
	Payload         string
}

type ActiveWorker struct {
	task    *Task
	process *Process
}

type FinishedTask struct {
	Task     *Task
	Response string
	IsError  bool
}

type ProcessTask struct {
	Task         *Task
	Process      *Process
	FinishedTask *FinishedTask
}
