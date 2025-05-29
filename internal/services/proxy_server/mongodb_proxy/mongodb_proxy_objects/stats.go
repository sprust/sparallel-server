package mongodb_proxy_objects

type OperationsStats struct {
	Running  int
	Finished int
	Timeouts int
}

type ServiceStats struct {
	InsertOne OperationsStats
	UpdateOne OperationsStats
	Aggregate OperationsStats
}
