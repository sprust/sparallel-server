package rpc_manager

type SleepArgs struct {
	Message string
}

type SleepResult struct {
	Answer string
}

type WakeUpArgs struct {
	Message string
}

type WakeUpResult struct {
	Answer string
}

type StopArgs struct {
	Message string
}

type StopResult struct {
	Answer string
}

type StatsArgs struct {
	Message string
}

type StatsResult struct {
	Json string
}
