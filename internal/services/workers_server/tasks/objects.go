package tasks

type Tasks struct {
	waiting  *SubTasks
	finished *SubTasks
}

type Group struct {
	uuid        string
	unixTimeout int
	tasks       map[string]*Task // map[TaskUuid]
}

func (g *Group) IsTimeout() bool {
	return isTimeout(g.unixTimeout, 5)
}

type Task struct {
	GroupUuid   string
	TaskUuid    string
	UnixTimeout int
	Payload     string
	IsFinished  bool
	Response    string
	IsError     bool
}

func (t *Task) IsTimeout() bool {
	return isTimeout(t.UnixTimeout, 5)
}
