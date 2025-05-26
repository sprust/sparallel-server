package operations

import (
	"context"
	"github.com/google/uuid"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/connections"
	"sparallel_server/internal/services/proxy_server/mongodb_proxy/objects"
	"sync"
)

type Operations[T OperationInterface] struct {
	mutex       sync.Mutex
	connections *connections.Connections
	running     map[string]T
	finished    map[string]T
}

func NewOperations[T OperationInterface](connections *connections.Connections) *Operations[T] {
	return &Operations[T]{
		connections: connections,
		running:     make(map[string]T),
		finished:    make(map[string]T),
	}
}

func (o *Operations[T]) Add(
	ctx context.Context,
	connName string,
	dbName string,
	collName string,
	operation T,
) *objects.RunningOperation {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	opUuid := uuid.New().String()

	o.running[opUuid] = operation

	go func() {
		coll, err := o.connections.Get(connName, dbName, collName)

		if err != nil {
			operation.Error(err)
		} else {
			operation.Execute(ctx, coll)

			o.mutex.Lock()
			defer o.mutex.Unlock()

			delete(o.running, opUuid)
			o.finished[opUuid] = operation
		}
	}()

	return &objects.RunningOperation{Uuid: opUuid}
}

func (o *Operations[T]) Pull(uuid string) OperationInterface {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	operation, exists := o.running[uuid]

	if exists {
		return operation
	}

	operation, exists = o.finished[uuid]

	if exists {
		return operation
	}

	panic("logic error: operation is not running or finished")
}
