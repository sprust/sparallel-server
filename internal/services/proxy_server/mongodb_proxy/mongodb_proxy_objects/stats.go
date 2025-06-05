package mongodb_proxy_objects

type OperationsStats struct {
	Items    int
	Timeouts int
}

type ServiceStats struct {
	InsertOne OperationsStats
	UpdateOne OperationsStats
	Aggregate OperationsStats
}
