package workers_server

type WorkersServerStats struct {
	Workers StatWorkers
	Tasks   StatTasks
}

type StatWorkers struct {
	Count        int
	FreeCount    int
	BusyCount    int
	LoadPercent  int
	AddedCount   int
	TookCount    int
	FreedCount   int
	DeletedCount int
}

type StatTasks struct {
	WaitingCount       int
	FinishedCount      int
	AddedTotalCount    int
	ReAddedTotalCount  int
	TookTotalCount     int
	FinishedTotalCount int
	SuccessTotalCount  int
	ErrorTotalCount    int
	TimeoutTotalCount  int
}
