package rpc_manager

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
