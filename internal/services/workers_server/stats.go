package workers_server

type WorkersServerStats struct {
	Workers StatWorkers
	Tasks   StatTasks
}

type StatWorkers struct {
	Count       int
	FreeCount   int
	BusyCount   int
	LoadPercent int
}

type StatTasks struct {
	WaitingCount  int
	FinishedCount int
}
